package market

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

func TestRefreshAssetPrice(t *testing.T) {
	assetID := uuid.New()
	userID := uuid.New()

	repoFor := func(asset Asset) *fakeRepository {
		return new(fakeRepository{
			getAssetByID: func(_ context.Context, id uuid.UUID) (Asset, error) {
				if id != assetID {
					t.Errorf("assetID = %s, want %s", id, assetID)
				}

				return asset, nil
			},
		})
	}

	stock := Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}

	quoting := func(price string, source ProviderID) *fakePriceProvider {
		return new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: price, Source: source}, nil
			},
		})
	}

	failing := func(provider ProviderID, sentinel error) *fakePriceProvider {
		return new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{}, marketdata.Errorf(provider, "user-key", sentinel, "provider said no")
			},
		})
	}

	t.Run("the price is fetched with the named key and stored against the user", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), quoting("190.55", Finnhub))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		got, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if err != nil {
			t.Fatalf("RefreshAssetPrice: %v", err)
		}
		if got.Price.String() != mustUSD(t, "190.55").String() {
			t.Errorf("Price = %s, want 190.55", got.Price)
		}
		if got.Source != Finnhub {
			t.Errorf("Source = %q, want finnhub", got.Source)
		}

		stored, ok := f.creds.priceOf(userID, assetID)
		if !ok {
			t.Fatal("the price was not stored against the user")
		}
		if stored.String() != mustUSD(t, "190.55").String() {
			t.Errorf("stored price = %s, want 190.55", stored)
		}
	})

	// The other half of a badge that tells the truth. Only a failure ever
	// reached SetCredentialStatus, so a key flagged on Tuesday stayed flagged
	// however many prices it fetched on Wednesday — the user had to notice and
	// press «Verificar» to clear a label that had stopped being true. Pressing
	// «Actualizar precio» is the most likely moment for that to be disproved.
	t.Run("a call that works retires a verdict a later call disproved", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), quoting("190.55", Finnhub))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		if err := f.creds.SetCredentialStatus(context.Background(), userID, Finnhub, CredentialRateLimited, "yesterday's burst"); err != nil {
			t.Fatalf("SetCredentialStatus: %v", err)
		}

		if _, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub); err != nil {
			t.Fatalf("RefreshAssetPrice: %v", err)
		}

		if got := f.creds.statusOf(userID, Finnhub); got != CredentialActive {
			t.Errorf("credential status = %q, want it back to active", got)
		}
	})

	// The property that makes this endpoint different from the sync: the chain
	// is built from one key, so it cannot fall through to a provider the caller
	// did not choose and report its price under the chosen one's name.
	t.Run("only the named provider's key is handed to the factory", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), quoting("190.55", AlphaVantage))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")
		f.creds.seed(t, f.ring, userID, AlphaVantage, "user-alpha-key")

		if _, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, AlphaVantage); err != nil {
			t.Fatalf("RefreshAssetPrice: %v", err)
		}

		if len(f.factory.gotCreds) != 1 {
			t.Fatalf("factory got %d credentials, want exactly 1", len(f.factory.gotCreds))
		}
		if f.factory.gotCreds[0].Provider != AlphaVantage {
			t.Errorf("provider = %q, want alphavantage", f.factory.gotCreds[0].Provider)
		}
		if f.factory.gotCreds[0].APIKey != "user-alpha-key" {
			t.Error("the factory was handed a key other than the one for the named provider")
		}
	})

	t.Run("a type no provider quotes is answered as such, not as a provider failure", func(t *testing.T) {
		cash := Asset{ID: assetID, Ticker: "COP", AssetType: Cash, Currency: money.COP}

		f := newBYOFixture(t, repoFor(cash), quoting("1", Finnhub))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if !errors.Is(err, ErrAssetNotQuotable) {
			t.Fatalf("err = %v, want ErrAssetNotQuotable", err)
		}
	})

	t.Run("a symbol the provider does not carry is about the asset, and leaves the key alone", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), failing(Finnhub, marketdata.ErrUnsupported))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if !errors.Is(err, ErrAssetNotCovered) {
			t.Fatalf("err = %v, want ErrAssetNotCovered", err)
		}
		if got := f.creds.statusOf(userID, Finnhub); got != CredentialActive {
			t.Errorf("credential status = %q, want it left active", got)
		}
	})

	t.Run("a spent quota is reported as such and marks the key rate limited", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), failing(Finnhub, marketdata.ErrRateLimited))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if !errors.Is(err, ErrProviderQuotaSpent) {
			t.Fatalf("err = %v, want ErrProviderQuotaSpent", err)
		}
		if got := f.creds.statusOf(userID, Finnhub); got != CredentialRateLimited {
			t.Errorf("credential status = %q, want rate_limited", got)
		}
	})

	// The bug this exists to pin: Alpha Vantage answers a burst with "please
	// consider spreading out your free API requests more sparingly", which is
	// about our request rate — and our IP, since a different key from the same
	// host trips it too. Recording it as a spent quota left a key that works a
	// second later wearing a standing "no quota" badge in the UI.
	t.Run("a throttle is reported as such and leaves the key alone", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), failing(Finnhub, marketdata.ErrThrottled))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if !errors.Is(err, ErrProviderThrottled) {
			t.Fatalf("err = %v, want ErrProviderThrottled", err)
		}
		if errors.Is(err, ErrProviderQuotaSpent) {
			t.Error("a throttle must not be reported as a spent quota: the remedies differ")
		}
		if got := f.creds.statusOf(userID, Finnhub); got != CredentialActive {
			t.Errorf("credential status = %q, want it left active", got)
		}
	})

	t.Run("a rejected key is reported and recorded, as the sync would", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), failing(Finnhub, marketdata.ErrUnauthorized))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, Finnhub)
		if !errors.Is(err, ErrInvalidAPIKey) {
			t.Fatalf("err = %v, want ErrInvalidAPIKey", err)
		}
		if got := f.creds.statusOf(userID, Finnhub); got != CredentialInvalid {
			t.Errorf("credential status = %q, want invalid", got)
		}
	})

	t.Run("a provider the user has no key for is a not-found, not a chain built from the others", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), quoting("190.55", Finnhub))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, AlphaVantage)
		if !errors.Is(err, ErrCredentialNotFound) {
			t.Fatalf("err = %v, want ErrCredentialNotFound", err)
		}
	})

	t.Run("an unknown provider is rejected before any key is opened", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(stock), quoting("190.55", Finnhub))
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		_, err := f.svc.RefreshAssetPrice(context.Background(), userID, assetID, ProviderID("yahoo"))
		if !errors.Is(err, ErrInvalidProvider) {
			t.Fatalf("err = %v, want ErrInvalidProvider", err)
		}
		if f.factory.gotCreds != nil {
			t.Error("the factory was called for a provider that does not exist")
		}
	})
}
