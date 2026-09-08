package market

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// providerErr builds the kind of error the real clients return: a scrubbed
// message plus an optional sentinel. A nil sentinel is what a timeout, a 5xx or
// an undecodable body produce, and telling that case apart from a rejected key
// is what most of this file is about.
func providerErr(provider ProviderID, sentinel error, msg string) error {
	return marketdata.Errorf(provider, "the-secret-key", sentinel, "%s", msg)
}

func quoteFailing(err error) *fakePriceProvider {
	return new(fakePriceProvider{
		fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
			return marketdata.QuoteResult{}, err
		},
	})
}

func quoteOK() *fakePriceProvider {
	return new(fakePriceProvider{
		fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
			return marketdata.QuoteResult{Price: "190.55", Source: Finnhub}, nil
		},
	})
}

func TestSaveCredential(t *testing.T) {
	userID := uuid.New()

	t.Run("a key the provider accepts is stored active", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		cred, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "a-real-looking-key")
		if err != nil {
			t.Fatalf("SaveCredential: %v", err)
		}

		if cred.Status != CredentialActive {
			t.Errorf("status = %q, want %q", cred.Status, CredentialActive)
		}
		if cred.Last4 != "-key" {
			t.Errorf("last4 = %q, want %q", cred.Last4, "-key")
		}
	})

	t.Run("a key the provider rejects is not stored at all", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, marketdata.ErrUnauthorized, "finnhub: status 401")))

		_, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "wrong-key")
		if !errors.Is(err, ErrInvalidAPIKey) {
			t.Fatalf("err = %v, want ErrInvalidAPIKey", err)
		}

		creds, _ := f.svc.ListCredentials(context.Background(), userID)
		if len(creds) != 0 {
			t.Fatalf("stored %d credentials, want none", len(creds))
		}
	})

	// The regression this file exists for: a provider that cannot be reached
	// says nothing about the key. Reporting it as a rejection told users their
	// key was bad during somebody else's outage.
	t.Run("an unreachable provider is not a rejected key", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, nil, "finnhub: http get quote: dial tcp: i/o timeout")))

		_, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "a-real-looking-key")
		if !errors.Is(err, ErrProviderUnavailable) {
			t.Fatalf("err = %v, want ErrProviderUnavailable", err)
		}
		if errors.Is(err, ErrInvalidAPIKey) {
			t.Error("an unreachable provider must not report the key as invalid")
		}
	})

	t.Run("a key whose quota is spent is still stored, flagged rate_limited", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(AlphaVantage, marketdata.ErrRateLimited, "alphavantage: note")))

		// Refusing the save here would mean a user who hit their daily limit
		// could not configure a key that works perfectly well.
		cred, err := f.svc.SaveCredential(context.Background(), userID, AlphaVantage, "quota-spent-key")
		if err != nil {
			t.Fatalf("SaveCredential: %v", err)
		}
		if cred.Status != CredentialRateLimited {
			t.Errorf("status = %q, want %q", cred.Status, CredentialRateLimited)
		}
	})

	t.Run("a symbol the provider does not cover still proves the key", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, marketdata.ErrUnsupported, "finnhub: no data")))

		cred, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "a-real-looking-key")
		if err != nil {
			t.Fatalf("SaveCredential: %v", err)
		}
		if cred.Status != CredentialActive {
			t.Errorf("status = %q, want %q", cred.Status, CredentialActive)
		}
	})

	t.Run("an unknown provider is rejected before any request", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		if _, err := f.svc.SaveCredential(context.Background(), userID, ProviderID("nasdaq"), "key"); !errors.Is(err, ErrInvalidProvider) {
			t.Fatalf("err = %v, want ErrInvalidProvider", err)
		}
	})

	t.Run("the stored key round-trips through the keyring", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		if _, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "round-trip-key"); err != nil {
			t.Fatalf("SaveCredential: %v", err)
		}

		plain, err := f.svc.openCredential(context.Background(), userID, Finnhub)
		if err != nil {
			t.Fatalf("openCredential: %v", err)
		}
		if string(plain) != "round-trip-key" {
			t.Errorf("opened %q, want the key that was saved", plain)
		}
	})

	t.Run("a sealed key cannot be opened under another user", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		if _, err := f.svc.SaveCredential(context.Background(), userID, Finnhub, "round-trip-key"); err != nil {
			t.Fatalf("SaveCredential: %v", err)
		}

		// The AAD binds the ciphertext to its owner, so moving the row is not a
		// way to read somebody else's key.
		sealed, err := f.creds.GetSealedCredential(context.Background(), userID, Finnhub)
		if err != nil {
			t.Fatalf("GetSealedCredential: %v", err)
		}

		if _, err := f.ring.Open(sealed.Sealed, credentialAAD(uuid.New().String(), string(Finnhub))); err == nil {
			t.Fatal("opened another user's ciphertext; the AAD binding is not holding")
		}
	})
}

func TestVerifyCredential(t *testing.T) {
	userID := uuid.New()

	t.Run("a working key is marked active", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		cred, err := f.svc.VerifyCredential(context.Background(), userID, Finnhub)
		if err != nil {
			t.Fatalf("VerifyCredential: %v", err)
		}
		if cred.Status != CredentialActive {
			t.Errorf("status = %q, want %q", cred.Status, CredentialActive)
		}
		if cred.LastVerifiedAt == nil {
			t.Error("lastVerifiedAt not recorded")
		}
	})

	t.Run("a revoked key is marked invalid", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, marketdata.ErrUnauthorized, "finnhub: status 401")))
		f.creds.seed(t, f.ring, userID, Finnhub, "revoked-key")

		if _, err := f.svc.VerifyCredential(context.Background(), userID, Finnhub); err != nil {
			t.Fatalf("VerifyCredential: %v", err)
		}

		if got := f.creds.statusOf(userID, Finnhub); got != CredentialInvalid {
			t.Errorf("status = %q, want %q", got, CredentialInvalid)
		}
	})

	// Same reasoning as the outage below, one step milder: a throttle is about
	// how fast we asked — and, since Alpha Vantage's burst limit follows our IP,
	// about our own concurrency — never about the key. Stamping 'rate_limited'
	// here left a working key showing "no quota" until somebody re-verified it.
	t.Run("a throttled provider leaves the stored status untouched", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(AlphaVantage, marketdata.ErrThrottled, "alphavantage: quote: please consider spreading out your free API requests more sparingly (1 request per second)")))
		f.creds.seed(t, f.ring, userID, AlphaVantage, "perfectly-good-key")

		_, err := f.svc.VerifyCredential(context.Background(), userID, AlphaVantage)
		if !errors.Is(err, ErrProviderThrottled) {
			t.Fatalf("err = %v, want ErrProviderThrottled", err)
		}

		if got := f.creds.statusOf(userID, AlphaVantage); got != CredentialActive {
			t.Fatalf("status = %q, want it left at %q", got, CredentialActive)
		}
	})

	// An outage must leave the row alone. Writing 'invalid' here takes the key
	// out of GetSealedCredentials and UsersWithCredentials for good, so one bad
	// afternoon at the provider would silently retire a key that still works.
	t.Run("an unreachable provider leaves the stored status untouched", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteFailing(providerErr(Finnhub, nil, "finnhub: http get quote: connection refused")))
		f.creds.seed(t, f.ring, userID, Finnhub, "still-good-key")

		_, err := f.svc.VerifyCredential(context.Background(), userID, Finnhub)
		if !errors.Is(err, ErrProviderUnavailable) {
			t.Fatalf("err = %v, want ErrProviderUnavailable", err)
		}

		if got := f.creds.statusOf(userID, Finnhub); got != CredentialActive {
			t.Fatalf("status = %q, want it left at %q", got, CredentialActive)
		}

		// And the key is still one the sync would use.
		sealed, err := f.creds.GetSealedCredentials(context.Background(), userID)
		if err != nil || len(sealed) != 1 {
			t.Fatalf("GetSealedCredentials = %d keys (%v), want the key to remain usable", len(sealed), err)
		}
	})

	t.Run("verifying a provider the user has not configured is a not-found", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		if _, err := f.svc.VerifyCredential(context.Background(), userID, Finnhub); !errors.Is(err, ErrCredentialNotFound) {
			t.Fatalf("err = %v, want ErrCredentialNotFound", err)
		}
	})
}

func TestDeleteCredential(t *testing.T) {
	userID := uuid.New()

	t.Run("deleting removes the key from the sync path", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		if err := f.svc.DeleteCredential(context.Background(), userID, Finnhub); err != nil {
			t.Fatalf("DeleteCredential: %v", err)
		}

		if _, _, err := f.svc.providerFor(context.Background(), userID); !errors.Is(err, ErrNoCredentials) {
			t.Fatalf("err = %v, want ErrNoCredentials once the last key is gone", err)
		}
	})

	t.Run("deleting a key that is not there is a not-found", func(t *testing.T) {
		f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

		if err := f.svc.DeleteCredential(context.Background(), userID, Finnhub); !errors.Is(err, ErrCredentialNotFound) {
			t.Fatalf("err = %v, want ErrCredentialNotFound", err)
		}
	})
}

func TestProviderForSkipsUnreadableKeys(t *testing.T) {
	userID := uuid.New()
	f := newBYOFixture(t, new(fakeRepository{}), quoteOK())

	f.creds.seed(t, f.ring, userID, Finnhub, "readable-key")

	// A row sealed under a keyring this service does not hold — what a retired
	// KEK version looks like. It must not sink the key that still opens.
	other := testKeyring()
	sealed, err := other.Seal([]byte("unreadable"), credentialAAD(userID.String(), string(AlphaVantage)))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	f.creds.sealed[userID] = append(f.creds.sealed[userID], sealedCredential{Provider: AlphaVantage, Sealed: sealed})

	if _, _, err := f.svc.providerFor(context.Background(), userID); err != nil {
		t.Fatalf("providerFor: %v", err)
	}

	if len(f.factory.gotCreds) != 1 || f.factory.gotCreds[0].Provider != Finnhub {
		t.Fatalf("handed over %+v, want only the key that opened", f.factory.gotCreds)
	}
}

func TestClassifyProbe(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus CredentialStatus
		wantErr    error
	}{
		{"success", nil, CredentialActive, nil},
		{"rejected key", providerErr(Finnhub, marketdata.ErrUnauthorized, "401"), CredentialInvalid, ErrInvalidAPIKey},
		{"spent quota", providerErr(Finnhub, marketdata.ErrRateLimited, "429"), CredentialRateLimited, nil},
		{"uncovered symbol", providerErr(Finnhub, marketdata.ErrUnsupported, "no data"), CredentialActive, nil},
		{"transport failure", providerErr(Finnhub, nil, "i/o timeout"), "", ErrProviderUnavailable},
		// An error that never went through a provider client carries no verdict
		// at all, so it cannot be read as one about the key either.
		{"unattributed failure", errors.New("something else"), "", ErrProviderUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := classifyProbe(tt.err)

			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
