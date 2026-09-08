package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// verifySymbol is a widely covered ticker used to prove a key works before it
// is stored. It costs one request against the user's own quota, which is worth
// it: the alternative is a key that looks saved and silently produces nothing
// until the next morning's sync.
const verifySymbol = "AAPL"

// ListCredentials returns what a user may see about their own keys: never the
// key, never the ciphertext.
func (s *service) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	return s.repo.ListCredentials(ctx, userID)
}

// SaveCredential verifies an API key against its provider and, only if the
// provider accepts it, seals and stores it.
//
// The plaintext key lives in this function and in the provider client it builds,
// and nowhere else: it is never logged, never returned, and never written
// unsealed.
func (s *service) SaveCredential(ctx context.Context, userID uuid.UUID, provider ProviderID, apiKey string) (Credential, error) {
	if !provider.IsValid() {
		return Credential{}, ErrInvalidProvider
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Credential{}, ErrInvalidAPIKey
	}

	// A key whose quota is spent is still a good key, so probe reports the status
	// to store rather than refusing the save: telling the user to come back
	// tomorrow to configure a key that works today would be absurd.
	status, err := s.probe(ctx, provider, apiKey)
	if err != nil {
		return Credential{}, err
	}

	sealed, err := s.seal(userID, provider, apiKey)
	if err != nil {
		return Credential{}, err
	}

	return s.repo.UpsertCredential(ctx, userID, sealed, last4(apiKey), status, new(time.Now().UTC()))
}

func (s *service) DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	if !provider.IsValid() {
		return ErrInvalidProvider
	}

	return s.repo.DeleteCredential(ctx, userID, provider)
}

// VerifyCredential re-checks a stored key and records the verdict, so a key
// that was revoked at the provider shows up as invalid in the UI instead of
// just producing no prices.
func (s *service) VerifyCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (Credential, error) {
	if !provider.IsValid() {
		return Credential{}, ErrInvalidProvider
	}

	apiKey, err := s.openCredential(ctx, userID, provider)
	if err != nil {
		return Credential{}, err
	}
	defer secretbox.Zero(apiKey)

	status, probeErr := s.probe(ctx, provider, string(apiKey))
	if errors.Is(probeErr, ErrProviderUnavailable) || errors.Is(probeErr, ErrProviderThrottled) {
		// The row is left exactly as it was. An outage — or a throttle, which is
		// about our request rate and our IP — is a statement about the provider,
		// not about the key, and writing a verdict here would either take the key
		// out of the sync queries until the user noticed and re-verified, or
		// leave it labelled as out of quota when it is not.
		return Credential{}, probeErr
	}

	statusErr := ""
	if probeErr != nil {
		statusErr = probeErr.Error()
	}

	if err := s.repo.SetCredentialStatus(ctx, userID, provider, status, statusErr); err != nil {
		return Credential{}, err
	}

	creds, err := s.repo.ListCredentials(ctx, userID)
	if err != nil {
		return Credential{}, err
	}
	for _, c := range creds {
		if c.Provider == provider {
			return c, nil
		}
	}

	return Credential{}, ErrCredentialNotFound
}

// probe spends one request to confirm the provider accepts the key, and reports
// the status that request justifies storing against it.
//
// A non-nil error means "do not record this outcome as a verdict on the key":
// either the provider rejected it (ErrInvalidAPIKey) or we never got an answer
// worth acting on (ErrProviderUnavailable).
func (s *service) probe(ctx context.Context, provider ProviderID, apiKey string) (CredentialStatus, error) {
	chain, err := s.providers.For([]marketdata.Credential{{Provider: provider, APIKey: apiKey}})
	if err != nil {
		return "", err
	}

	_, err = chain.FetchQuote(ctx, verifySymbol)

	return classifyProbe(err)
}

// classifyProbe turns a provider failure into the status to store.
//
// It reads the failure through marketdata.Verdicts rather than errors.Is for
// the same reason the sync job does (see recordProviderVerdict): the sentinel is
// what carries meaning, and a failure with no sentinel — a timeout, a 5xx, a
// body that would not decode — carries none. Treating that silence as "the
// provider rejected the key" is what used to retire working keys during an
// outage.
func classifyProbe(err error) (CredentialStatus, error) {
	if err == nil {
		return CredentialActive, nil
	}

	verdicts := marketdata.Verdicts(err)
	if len(verdicts) == 0 {
		return "", ErrProviderUnavailable
	}

	var unauthorized, rateLimited, throttled, unsupported bool

	for _, verdict := range verdicts {
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
		return CredentialInvalid, ErrInvalidAPIKey
	case rateLimited:
		// The key is fine, its quota is not. Stored as-is so the UI can say so,
		// and so tomorrow's sync tries it again instead of skipping it.
		return CredentialRateLimited, nil
	case throttled:
		// The provider asked us to slow down, which proves nothing either way:
		// the same key answers when the calls are spaced out, and the throttle
		// is shared across our keys because it follows our IP. Treated like an
		// outage — the row is left exactly as it was — because writing
		// 'rate_limited' here is what used to leave a working key wearing a
		// standing "no quota" badge.
		return "", ErrProviderThrottled
	case unsupported:
		// The provider answered and did not object to the key; it just has no
		// data for this symbol. That still proves the key works.
		return CredentialActive, nil
	default:
		return "", ErrProviderUnavailable
	}
}

// seal encrypts the key for storage, bound to the user and provider that own it.
func (s *service) seal(userID uuid.UUID, provider ProviderID, apiKey string) (sealedCredential, error) {
	if s.keyring == nil {
		return sealedCredential{}, ErrKeyEncryptionOff
	}

	sealed, err := s.keyring.Seal([]byte(apiKey), credentialAAD(userID.String(), string(provider)))
	if err != nil {
		return sealedCredential{}, fmt.Errorf("seal credential: %w", err)
	}

	return sealedCredential{Provider: provider, Sealed: sealed}, nil
}

// openCredential decrypts one stored key. The caller owns the returned bytes
// and must Zero them.
func (s *service) openCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) ([]byte, error) {
	if s.keyring == nil {
		return nil, ErrKeyEncryptionOff
	}

	stored, err := s.repo.GetSealedCredential(ctx, userID, provider)
	if err != nil {
		return nil, err
	}

	return s.keyring.Open(stored.Sealed, credentialAAD(userID.String(), string(provider)))
}

// providerFor assembles the chain for one user from their stored keys. The
// plaintext keys are zeroed before returning: the clients hold their own copy
// for the duration of the sync, and nothing else needs them.
//
// It also reports the pace to leave between calls, taken from the slowest
// provider the user actually configured.
func (s *service) providerFor(ctx context.Context, userID uuid.UUID) (marketdata.Provider, time.Duration, error) {
	if s.keyring == nil {
		return nil, 0, ErrKeyEncryptionOff
	}

	stored, err := s.repo.GetSealedCredentials(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if len(stored) == 0 {
		return nil, 0, ErrNoCredentials
	}

	creds := make([]marketdata.Credential, 0, len(stored))
	pace := finnhubPace

	for _, sc := range stored {
		plain, err := s.keyring.Open(sc.Sealed, credentialAAD(userID.String(), string(sc.Provider)))
		if err != nil {
			// One unreadable row (e.g. its KEK was retired) must not sink the
			// keys that still work.
			s.log.Error(ctx, "cannot open stored credential", logger.Err(err), logger.Str("provider", string(sc.Provider)))
			continue
		}

		creds = append(creds, marketdata.Credential{Provider: sc.Provider, APIKey: string(plain)})
		secretbox.Zero(plain)

		if sc.Provider == AlphaVantage {
			pace = alphaVantagePace
		}
	}

	chain, err := s.providers.For(creds)
	if err != nil {
		return nil, 0, err
	}

	return chain, pace, nil
}
