package market

import (
	"context"
	"errors"
	"strings"
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// byoFixture wires a service with a real keyring, an in-memory credential store
// and a stub provider, which is what every BYO-key sync scenario needs.
type byoFixture struct {
	svc     *service
	creds   *credentialStore
	factory *fakeFactory
	ring    *secretbox.Keyring
}

func newBYOFixture(t *testing.T, repo *fakeRepository, provider marketdata.Provider) *byoFixture {
	t.Helper()

	ring := testKeyring()
	creds := newCredentialStore()
	repo.creds = creds
	factory := new(fakeFactory{provider: provider})

	return new(byoFixture{
		svc:     newService(repo, newMemStorage(), factory, nil, ring, logger.Noop()),
		creds:   creds,
		factory: factory,
		ring:    ring,
	})
}

func TestSyncAssetsForUser(t *testing.T) {
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

	// A key that just filled the whole portfolio is not rate limited, whatever
	// yesterday's run wrote about it. Without this the daily sync could never
	// clear its own verdict: only failures reached SetCredentialStatus.
	t.Run("a successful run clears the verdict left by an earlier one", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "190.55", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		if err := f.creds.SetCredentialStatus(context.Background(), userID, Finnhub, CredentialRateLimited, "an earlier burst"); err != nil {
			t.Fatalf("SetCredentialStatus: %v", err)
		}

		if _, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID}); len(errs) > 0 {
			t.Fatalf("SyncAssetsForUser: %v", errs)
		}

		if got := f.creds.statusOf(userID, Finnhub); got != CredentialActive {
			t.Errorf("credential status = %q, want it back to active", got)
		}
	})

	t.Run("a stock price is fetched with the user's key and stored against them", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(_ context.Context, symbol string) (marketdata.QuoteResult, error) {
				if symbol != "AAPL" {
					t.Errorf("symbol = %q, want AAPL", symbol)
				}

				return marketdata.QuoteResult{Price: "190.55", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		got, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID})
		if len(errs) > 0 {
			t.Fatalf("SyncAssetsForUser: %v", errs)
		}
		if len(got) != 1 || got[0].Price.String() != mustUSD(t, "190.55").String() {
			t.Fatalf("results = %+v, want one price of 190.55", got)
		}
		if got[0].Source != Finnhub {
			t.Errorf("Source = %q, want finnhub", got[0].Source)
		}

		// Stored against this user, not on the shared catalog row.
		stored, ok := f.creds.priceOf(userID, assetID)
		if !ok {
			t.Fatal("the price was not stored against the user")
		}
		if stored.String() != mustUSD(t, "190.55").String() {
			t.Errorf("stored price = %s, want 190.55", stored)
		}
	})

	t.Run("the user's own key is the one handed to the provider factory", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "1", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "the-users-own-key")

		if _, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID}); len(errs) > 0 {
			t.Fatalf("SyncAssetsForUser: %v", errs)
		}

		if len(f.factory.gotCreds) != 1 {
			t.Fatalf("factory got %d credentials, want 1", len(f.factory.gotCreds))
		}
		if f.factory.gotCreds[0].APIKey != "the-users-own-key" {
			t.Errorf("factory got the wrong key back from the sealed store")
		}
	})

	t.Run("a user with no key gets ErrNoCredentials, not a provider call", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				t.Fatal("the provider must not be called without a credential")

				return marketdata.QuoteResult{}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}), provider)

		_, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID})
		if len(errs) != 1 || !errors.Is(errs[0], marketdata.ErrNoCredentials) {
			t.Fatalf("errs = %v, want ErrNoCredentials", errs)
		}
	})

	t.Run("one user's key cannot open another user's stored credential", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}), new(fakePriceProvider{}))

		other := uuid.New()
		f.creds.seed(t, f.ring, userID, Finnhub, "alices-key")

		// Move Alice's sealed row onto Bob, exactly as a database attacker with
		// write access would. The AAD binding must defeat it.
		f.creds.mu.Lock()
		f.creds.sealed[other] = f.creds.sealed[userID]
		f.creds.mu.Unlock()

		_, errs := f.svc.SyncAssetsForUser(context.Background(), other, []uuid.UUID{assetID})
		if len(errs) != 1 || !errors.Is(errs[0], marketdata.ErrNoCredentials) {
			t.Fatalf("errs = %v, want ErrNoCredentials — the moved row must not open", errs)
		}
	})

	t.Run("crypto prices come from an exchange rate on the split ticker", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchExchangeRate: func(_ context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error) {
				if from != money.EUR || to != money.USD {
					t.Errorf("pair = %s/%s, want EUR/USD", from, to)
				}

				return marketdata.ExchangeRateResult{Rate: "64000.00", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "EUR-USD", AssetType: Crypto, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		got, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID})
		if len(errs) > 0 {
			t.Fatalf("SyncAssetsForUser: %v", errs)
		}
		if len(got) != 1 || got[0].Price.String() != mustUSD(t, "64000.00").String() {
			t.Fatalf("results = %+v, want one price of 64000.00", got)
		}
	})

	// The pair a crypto ticker actually names is the one FetchExchangeRate
	// cannot carry: money.Currency is ISO 4217, and no crypto is in that table.
	// Until the provider interface takes the leg as text again, BTC-USD has to
	// fail saying so rather than quietly asking the feed for XXX/USD.
	t.Run("a crypto ticker outside ISO 4217 fails naming the leg", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchExchangeRate: func(context.Context, money.Currency, money.Currency) (marketdata.ExchangeRateResult, error) {
				t.Fatal("an unrepresentable pair must not reach the provider")
				return marketdata.ExchangeRateResult{}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "BTC-USD", AssetType: Crypto, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		_, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID})
		if len(errs) != 1 {
			t.Fatalf("errs = %v, want one error", errs)
		}
		if !strings.Contains(errs[0].Error(), "BTC") {
			t.Errorf("err = %v, want it to name the leg it could not express", errs[0])
		}
	})

	t.Run("unquotable asset types are skipped, not failed", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				t.Fatal("cash must not reach the provider")

				return marketdata.QuoteResult{}, nil
			},
		})

		f := newBYOFixture(t, repoFor(Asset{ID: assetID, Ticker: "CASH", AssetType: Cash, Currency: money.USD}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		got, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID})
		if len(errs) > 0 || len(got) != 0 {
			t.Fatalf("got %+v / %v, want both empty", got, errs)
		}
	})

	t.Run("nothing to sync makes no provider call at all", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), new(fakePriceProvider{}))

		got, errs := f.svc.SyncAssetsForUser(context.Background(), userID, nil)
		if len(got) != 0 || len(errs) != 0 {
			t.Fatalf("got %+v / %v, want both empty", got, errs)
		}
	})
}

// A rejected key is marked invalid so the daily job stops spending requests on
// it; a throttled one stays active so tomorrow tries again. Crucially, the
// verdict lands only on the provider that produced it.
func TestSyncRecordsTheProviderVerdictOnTheRightKey(t *testing.T) {
	assetID, userID := uuid.New(), uuid.New()

	repo := new(fakeRepository{
		getAssetByID: func(context.Context, uuid.UUID) (Asset, error) {
			return Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: money.USD}, nil
		},
	})

	// Finnhub rejects its key; Alpha Vantage is merely throttled. A fallback
	// chain joins both failures.
	provider := new(fakePriceProvider{
		fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
			return marketdata.QuoteResult{}, errors.Join(
				marketdata.Errorf(marketdata.Finnhub, "k1", marketdata.ErrUnauthorized, "finnhub: status 401"),
				marketdata.Errorf(marketdata.AlphaVantage, "k2", marketdata.ErrRateLimited, "alphavantage: quota"),
			)
		},
	})

	f := newBYOFixture(t, repo, provider)
	f.creds.seed(t, f.ring, userID, Finnhub, "k1")
	f.creds.seed(t, f.ring, userID, AlphaVantage, "k2")

	if _, errs := f.svc.SyncAssetsForUser(context.Background(), userID, []uuid.UUID{assetID}); len(errs) == 0 {
		t.Fatal("expected the sync to report the failure")
	}

	if got := f.creds.statusOf(userID, Finnhub); got != CredentialInvalid {
		t.Errorf("finnhub status = %q, want invalid", got)
	}
	if got := f.creds.statusOf(userID, AlphaVantage); got != CredentialRateLimited {
		t.Errorf("alphavantage status = %q, want rate_limited — a working key must not be demoted", got)
	}
}

func TestSyncRatesForUser(t *testing.T) {
	userID := uuid.New()

	t.Run("rates are stored against the user who fetched them", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchExchangeRate: func(_ context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error) {
				return marketdata.ExchangeRateResult{Rate: "1.09", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, new(fakeRepository{}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		got, errs := f.svc.SyncRatesForUser(context.Background(), userID, []CurrencyPair{{From: money.EUR, To: money.USD}})
		if len(errs) > 0 {
			t.Fatalf("SyncRatesForUser: %v", errs)
		}
		if len(got) != 1 || got[0].Rate.String() != "1.09" {
			t.Fatalf("results = %+v, want one rate of 1.09", got)
		}

		f.creds.mu.Lock()
		defer f.creds.mu.Unlock()
		if _, ok := f.creds.rates[userID.String()+"/EURUSD"]; !ok {
			t.Error("the rate was not stored against the user")
		}
	})

	t.Run("unparseable rates are rejected", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchExchangeRate: func(context.Context, money.Currency, money.Currency) (marketdata.ExchangeRateResult, error) {
				return marketdata.ExchangeRateResult{Rate: "not-a-number", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, new(fakeRepository{}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		got, errs := f.svc.SyncRatesForUser(context.Background(), userID, []CurrencyPair{{From: money.EUR, To: money.USD}})
		if len(got) != 0 || len(errs) != 1 {
			t.Fatalf("got %+v / %v, want no results and one error", got, errs)
		}
	})

	t.Run("a fetch failure is collected per pair", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchExchangeRate: func(context.Context, money.Currency, money.Currency) (marketdata.ExchangeRateResult, error) {
				return marketdata.ExchangeRateResult{}, errors.New("provider down")
			},
		})

		f := newBYOFixture(t, new(fakeRepository{}), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "key")

		pairs := []CurrencyPair{{money.EUR, money.USD}, {money.GBP, money.USD}, {money.USD, money.COP}}

		_, errs := f.svc.SyncRatesForUser(context.Background(), userID, pairs)
		if len(errs) != len(pairs) {
			t.Fatalf("errs = %d, want one per pair (%d)", len(errs), len(pairs))
		}
	})
}
