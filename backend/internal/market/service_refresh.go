package market

import (
	"context"
	"errors"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// RefreshAssetPrice re-quotes one asset with one named key of the caller's own
// and stores the answer as that user's price for it.
//
// It is the single-asset, single-provider counterpart of SyncAssetsForUser.
// That one walks every holding and lets the fallback chain decide who answers,
// which is right for a daily job and wrong for somebody looking at one
// position: the price on screen came from a provider that can be named, and
// re-asking *that* provider is the only way to tell a stale number apart from
// one the provider still stands behind.
func (s *service) RefreshAssetPrice(ctx context.Context, userID, assetID uuid.UUID, provider ProviderID) (UserAssetPrice, error) {
	if !provider.IsValid() {
		return UserAssetPrice{}, ErrInvalidProvider
	}

	chain, err := s.providerForOne(ctx, userID, provider)
	if err != nil {
		return UserAssetPrice{}, err
	}

	log := s.log.With(
		logger.Str("job", "asset_price_refresh"),
		logger.Str("userID", userID.String()),
		logger.Str("provider", string(provider)),
	)

	price, err := s.syncOneAsset(ctx, userID, assetID, chain, log)
	if err != nil {
		if errors.Is(err, errAssetTypeUnsupported) {
			// Cash, real estate and the rest. The sync skips these in silence,
			// which is right for a job over every holding; here it is the answer
			// to a question somebody asked, and it is about the asset, not the
			// key.
			return UserAssetPrice{}, ErrAssetNotQuotable
		}

		log.Error(ctx, "refresh asset price failed", logger.Err(err), logger.Str("assetID", assetID.String()))

		// Same bookkeeping the sync does: a key the provider just rejected is
		// marked here too, or it would keep looking healthy in settings until
		// the next morning's run.
		s.recordProviderVerdict(ctx, userID, err)

		return UserAssetPrice{}, classifyRefresh(err)
	}

	// The key just worked, so any badge it was wearing is out of date. This is
	// the button a user presses after fixing something, which makes it the most
	// likely place for a stale verdict to be disproved.
	s.clearStaleVerdicts(ctx, userID, map[ProviderID]bool{price.Source: true})

	return price, nil
}

// providerForOne builds a chain from exactly one of the user's keys.
//
// providerFor assembles every key they have and hands the lot to the fallback,
// which is what makes a sync resilient. Here that would be a bug: the caller
// picked a provider, and a chain that quietly fell through to the next one
// would store Alpha Vantage's price under Finnhub's name.
func (s *service) providerForOne(ctx context.Context, userID uuid.UUID, provider ProviderID) (marketdata.Provider, error) {
	if s.keyring == nil {
		return nil, ErrKeyEncryptionOff
	}

	apiKey, err := s.openCredential(ctx, userID, provider)
	if err != nil {
		return nil, err
	}
	defer secretbox.Zero(apiKey)

	return s.providers.For([]marketdata.Credential{{Provider: provider, APIKey: string(apiKey)}})
}

// classifyRefresh turns what the provider said into the domain error that says
// what the user can do about it.
//
// The sync path deliberately does not do this: it collects raw failures, records
// a verdict against the key and moves on to the next holding. Here the failure
// *is* the response, so it has to carry a status and a sentence — a spent quota
// is a 429 worth retrying tomorrow, an uncovered symbol is a 400 no retry will
// fix, and a rejected key is a trip to Ajustes.
//
// An error carrying no verdict — a timeout, a 5xx, an asset that is not in the
// catalog — is returned untouched, so it keeps whatever status it already had.
func classifyRefresh(err error) error {
	var unauthorized, rateLimited, throttled, unsupported bool

	for _, verdict := range marketdata.Verdicts(err) {
		switch {
		case errors.Is(verdict.Err, marketdata.ErrUnauthorized):
			unauthorized = true
		case errors.Is(verdict.Err, marketdata.ErrRateLimited):
			rateLimited = true
		case errors.Is(verdict.Err, marketdata.ErrThrottled):
			throttled = true
		case errors.Is(verdict.Err, marketdata.ErrUnsupported):
			unsupported = true
		}
	}

	switch {
	case unauthorized:
		return ErrInvalidAPIKey
	case rateLimited:
		return ErrProviderQuotaSpent
	case throttled:
		// Not the same answer as a spent quota, and not the same remedy: this one
		// clears in seconds.
		return ErrProviderThrottled
	case unsupported:
		return ErrAssetNotCovered
	default:
		return err
	}
}
