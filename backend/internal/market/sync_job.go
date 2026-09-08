package market

import (
	"context"
	"errors"
	"sync/atomic"

	"uuid"

	"golang.org/x/sync/errgroup"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// syncConcurrency bounds how many users are synced at once. Quotas are personal
// now, so users do not contend for one budget and can genuinely run in
// parallel; the cap is about our own outbound connections, not about the keys.
//
// It is no longer what protects us from a provider's per-IP burst limit, and it
// never could: four concurrent users are four calls leaving at once from one
// address whatever this number is, unless it is one. That limit belongs to the
// provider and is now enforced where it is known — the Alpha Vantage client
// paces every call this process makes, across all users (see its burstInterval).
const syncConcurrency = 4

// Holdings reports which assets a user actually owns, and which currency pairs
// their portfolios need converting between.
//
// It is an interface declared here and satisfied by the portfolio module, which
// is what keeps the module graph acyclic: portfolio already depends on market,
// so market must not import it back. The composition root supplies the
// implementation.
type Holdings interface {
	HeldAssetIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	RequiredCurrencyPairs(ctx context.Context, userID uuid.UUID) ([]CurrencyPair, error)
}

// SyncJob refreshes market data for every user who has configured a key, each
// with their own. It replaces the two global daily jobs, which no longer have a
// key to run under: the application holds no provider credentials.
type SyncJob struct {
	service  *service
	holdings Holdings
	log      logger.Logger
}

func NewSyncJob(service *service, holdings Holdings, log logger.Logger) *SyncJob {
	return new(SyncJob{
		service:  service,
		holdings: holdings,
		log:      log.With(logger.Str("scheduler", "market_sync")),
	})
}

func (j *SyncJob) Name() string { return "market-sync" }

func (j *SyncJob) Run(ctx context.Context) error {
	// The shared catalog needs no key, so it is refreshed once for everyone
	// before the per-user work starts.
	if errs := j.service.SeedDefaultAssets(ctx); len(errs) > 0 {
		j.log.Error(ctx, "seeding default assets reported failures", logger.Int("failed", len(errs)))
	}

	userIDs, err := j.service.repo.UsersWithCredentials(ctx)
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		j.log.Info(ctx, "no user has configured a market data key; nothing to sync")

		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(syncConcurrency)

	// failures counts users whose sync produced errors. One user's exhausted
	// quota must not abort everyone else's run, so each user's failures are
	// logged and counted rather than returned.
	var failures atomic.Int64

	for _, userID := range userIDs {
		group.Go(func() error {
			if j.syncUser(groupCtx, userID) {
				failures.Add(1)
			}

			// Never propagate: errgroup would cancel the siblings.
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	if n := failures.Load(); n > 0 {
		j.log.Error(ctx, "market sync completed with failures", logger.Int("failed_users", int(n)))

		return errors.New("market sync: some users failed to sync")
	}

	return nil
}

// syncUser runs one user's assets and rates. It reports whether anything failed
// rather than returning an error, because a single user's failure is not the
// job's failure.
func (j *SyncJob) syncUser(ctx context.Context, userID uuid.UUID) bool {
	log := j.log.With(logger.Str("userID", userID.String()))

	assetIDs, err := j.holdings.HeldAssetIDs(ctx, userID)
	if err != nil {
		log.Error(ctx, "cannot read holdings", logger.Err(err))

		return true
	}

	failed := false

	if _, errs := j.service.SyncAssetsForUser(ctx, userID, assetIDs); len(errs) > 0 {
		// A user with no key configured is an expected state, not a failure:
		// they simply see their holdings valued at cost. It is also the only
		// reason to stop here — without a key the rates cannot be fetched either.
		if errors.Is(errs[0], ErrNoCredentials) {
			return false
		}

		// Anything else is a failure of some assets, not of the user's run. The
		// rates still have to be refreshed: one ticker the provider does not
		// cover must not leave a multi-currency portfolio without conversions.
		log.Error(ctx, "asset sync reported failures", logger.Int("failed", len(errs)))
		failed = true
	}

	pairs, err := j.holdings.RequiredCurrencyPairs(ctx, userID)
	if err != nil {
		log.Error(ctx, "cannot read currency pairs", logger.Err(err))

		return true
	}

	if _, errs := j.service.SyncRatesForUser(ctx, userID, pairs); len(errs) > 0 {
		log.Error(ctx, "rate sync reported failures", logger.Int("failed", len(errs)))
		failed = true
	}

	return failed
}
