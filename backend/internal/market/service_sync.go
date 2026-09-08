package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// UserAssetPrice is a price fetched with one user's key. It is deliberately not
// an Asset: the price belongs to the user whose key paid for it and must never
// be handed to anybody else, so it never lands on the shared catalog row.
type UserAssetPrice struct {
	AssetID   uuid.UUID   `json:"assetId"`
	Ticker    string      `json:"ticker"`
	Price     money.Money `json:"price"`
	Source    ProviderID  `json:"source"`
	FetchedAt time.Time   `json:"fetchedAt"`
}

// UserExchangeRate is the same idea for a currency pair.
type UserExchangeRate struct {
	FromCurrency money.Currency  `json:"fromCurrency"`
	ToCurrency   money.Currency  `json:"toCurrency"`
	Rate         decimal.Decimal `json:"rate"`
	Source       ProviderID      `json:"source"`
	FetchedAt    time.Time       `json:"fetchedAt"`
}

// SyncAssetsForUser refreshes the given assets using the user's own keys and
// stores the results against that user.
//
// assetIDs is a parameter rather than something this service works out for
// itself: knowing which assets a user holds means reading portfolios, and
// market must not depend on portfolio (the dependency runs the other way, and
// app/arch_test.go pins it). The caller supplies the holdings.
func (s *service) SyncAssetsForUser(ctx context.Context, userID uuid.UUID, assetIDs []uuid.UUID) ([]UserAssetPrice, []error) {
	log := s.log.With(logger.Str("job", "asset_price"), logger.Str("userID", userID.String()))

	if len(assetIDs) == 0 {
		return nil, nil
	}

	chain, pace, err := s.providerFor(ctx, userID)
	if err != nil {
		return nil, []error{err}
	}

	results := make([]UserAssetPrice, 0, len(assetIDs))
	var errs []error
	// Who actually answered. Cleared once at the end rather than per price: a
	// hundred positions would otherwise mean a hundred UPDATEs saying the same
	// thing.
	answered := make(map[ProviderID]bool, len(SupportedProviders))

	for i, assetID := range assetIDs {
		// Space the calls out so a personal free-tier quota is not tripped by
		// our own burst. Only between calls, never before the first.
		if i > 0 {
			select {
			case <-ctx.Done():
				return results, errs
			case <-time.After(pace):
			}
		}

		price, err := s.syncOneAsset(ctx, userID, assetID, chain, log)
		if err != nil {
			if errors.Is(err, errAssetTypeUnsupported) {
				continue
			}

			log.Error(ctx, "sync asset failed", logger.Err(err), logger.Str("assetID", assetID.String()))
			errs = append(errs, err)
			s.recordProviderVerdict(ctx, userID, err)

			continue
		}

		results = append(results, price)
		answered[price.Source] = true
	}

	s.clearStaleVerdicts(ctx, userID, answered)

	return results, errs
}

// clearStaleVerdicts retires the badge of every key that just did its job.
//
// recordProviderVerdict is the other half and only a failure reaches it, so
// without this a key marked rate_limited stays marked however many prices it
// fetches afterwards — the user has to notice and press «Verificar» to clear a
// label that stopped being true days ago.
func (s *service) clearStaleVerdicts(ctx context.Context, userID uuid.UUID, answered map[ProviderID]bool) {
	for provider := range answered {
		if err := s.repo.MarkCredentialWorking(ctx, userID, provider); err != nil {
			// Not worth failing a run that fetched prices: the badge is stale,
			// the data is not.
			s.log.Error(ctx, "cannot clear credential status", logger.Err(err), logger.Str("provider", string(provider)))
		}
	}
}

// errAssetTypeUnsupported marks an asset the providers cannot quote (cash, real
// estate…). Not a failure, just nothing to do.
var errAssetTypeUnsupported = errors.New("asset type not quotable")

func (s *service) syncOneAsset(ctx context.Context, userID, assetID uuid.UUID, chain marketdata.Provider, log logger.Logger) (UserAssetPrice, error) {
	asset, err := s.repo.GetAssetByID(ctx, assetID)
	if err != nil {
		return UserAssetPrice{}, err
	}

	var priceStr string
	var source ProviderID

	switch asset.AssetType {
	case Stock, ETF, Bond:
		result, err := chain.FetchQuote(ctx, asset.Ticker)
		if err != nil {
			return UserAssetPrice{}, fmt.Errorf("fetch quote %q: %w", asset.Ticker, err)
		}
		priceStr, source = result.Price, result.Source

	case Crypto:
		base, quote, ok := strings.Cut(asset.Ticker, "-")
		if !ok {
			return UserAssetPrice{}, fmt.Errorf("cannot parse crypto ticker %q", asset.Ticker)
		}

		// FetchExchangeRate speaks money.Currency, which is ISO 4217 and
		// nothing else, so a leg the table does not hold — "BTC", "ETH" — has
		// no value to pass. Rejecting it here is what keeps that visible: a
		// discarded Scan error leaves the leg at XXX, and the provider is then
		// asked for XXX/USD and answers "no rate", blaming the feed for a
		// ticker this code could not express.
		var from, to money.Currency

		if err := from.Scan(base); err != nil {
			return UserAssetPrice{}, fmt.Errorf("crypto ticker %q: %q is not an ISO 4217 currency: %w", asset.Ticker, base, err)
		}

		if err := to.Scan(quote); err != nil {
			return UserAssetPrice{}, fmt.Errorf("crypto ticker %q: %q is not an ISO 4217 currency: %w", asset.Ticker, quote, err)
		}

		result, err := chain.FetchExchangeRate(ctx, from, to)
		if err != nil {
			return UserAssetPrice{}, fmt.Errorf("fetch exchange rate %q: %w", asset.Ticker, err)
		}

		priceStr, source = result.Rate, result.Source

	default:
		return UserAssetPrice{}, errAssetTypeUnsupported
	}

	price, err := money.NewMoneyFromString(priceStr, asset.Currency)
	if err != nil {
		return UserAssetPrice{}, fmt.Errorf("parse price for %q: %w", asset.Ticker, err)
	}

	fetchedAt := time.Now().UTC()
	if err := s.repo.UpsertUserAssetPrice(ctx, userID, asset.ID, price, asset.Currency, source, fetchedAt); err != nil {
		return UserAssetPrice{}, fmt.Errorf("persist price for %q: %w", asset.Ticker, err)
	}

	log.Info(ctx, "asset price updated", logger.Str("ticker", asset.Ticker), logger.Str("source", string(source)))

	return UserAssetPrice{
		AssetID:   asset.ID,
		Ticker:    asset.Ticker,
		Price:     price,
		Source:    source,
		FetchedAt: fetchedAt,
	}, nil
}

// SyncRatesForUser refreshes the given currency pairs with the user's own keys.
func (s *service) SyncRatesForUser(ctx context.Context, userID uuid.UUID, pairs []CurrencyPair) ([]UserExchangeRate, []error) {
	log := s.log.With(logger.Str("job", "exchange_rate"), logger.Str("userID", userID.String()))

	if len(pairs) == 0 {
		return nil, nil
	}

	chain, pace, err := s.providerFor(ctx, userID)
	if err != nil {
		return nil, []error{err}
	}

	results := make([]UserExchangeRate, 0, len(pairs))
	var errs []error
	answered := make(map[ProviderID]bool, len(SupportedProviders))

	for i, pair := range pairs {
		if i > 0 {
			select {
			case <-ctx.Done():
				return results, errs
			case <-time.After(pace):
			}
		}

		result, err := chain.FetchExchangeRate(ctx, pair.From, pair.To)
		if err != nil {
			log.Error(ctx, "fetch rate failed", logger.Err(err), logger.Str("from", pair.From.String()), logger.Str("to", pair.To.String()))
			errs = append(errs, err)
			s.recordProviderVerdict(ctx, userID, err)

			continue
		}

		rate, err := decimal.NewFromString(result.Rate)
		if err != nil {
			errs = append(errs, fmt.Errorf("parse rate %s/%s: %w", pair.From, pair.To, err))

			continue
		}

		if err := s.repo.UpsertUserExchangeRate(ctx, userID, pair.From, pair.To, rate, result.Source, result.FetchedAt); err != nil {
			errs = append(errs, fmt.Errorf("persist rate %s/%s: %w", pair.From, pair.To, err))

			continue
		}

		results = append(results, UserExchangeRate{
			FromCurrency: pair.From,
			ToCurrency:   pair.To,
			Rate:         rate,
			Source:       result.Source,
			FetchedAt:    result.FetchedAt,
		})
		answered[result.Source] = true
	}

	s.clearStaleVerdicts(ctx, userID, answered)

	return results, errs
}

// recordProviderVerdict marks a key invalid or rate-limited when its provider
// says so, so a revoked key surfaces in the UI instead of quietly yielding no
// prices every morning.
//
// The verdict is applied per provider, never to the whole chain: a fallback
// failure joins the errors of every key the user has, and one provider
// rejecting its key is no reason to demote a different key that still works.
func (s *service) recordProviderVerdict(ctx context.Context, userID uuid.UUID, err error) {
	for _, verdict := range marketdata.Verdicts(err) {
		var status CredentialStatus

		switch {
		case errors.Is(verdict.Err, marketdata.ErrUnauthorized):
			status = CredentialInvalid
		case errors.Is(verdict.Err, marketdata.ErrRateLimited):
			status = CredentialRateLimited
		default:
			// Transport failures and uncovered symbols say nothing about the key.
			continue
		}

		if setErr := s.repo.SetCredentialStatus(ctx, userID, verdict.Provider, status, verdict.Message); setErr != nil {
			s.log.Error(ctx, "cannot record credential status", logger.Err(setErr), logger.Str("provider", string(verdict.Provider)))
		}
	}
}
