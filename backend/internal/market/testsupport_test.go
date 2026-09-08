package market

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly. The asset
// hooks default to a no-op success when unset, so seeding loops don't have to
// stub every collaborator call along the way.
type fakeRepository struct {
	Repository
	// creds backs the CredentialStore half. Nil in scenarios that do not touch
	// BYO-key, in which case those methods panic like any other unstubbed one.
	creds *credentialStore

	upsertExchangeRate       func(ctx context.Context, from, to money.Currency, rate decimal.Decimal, rateDate time.Time) (ExchangeRate, error)
	upsertPublicExchangeRate func(ctx context.Context, from, to money.Currency, rate decimal.Decimal, rateDate time.Time, source ProviderID) (ExchangeRate, error)
	getExchangeRates         func(ctx context.Context, offset, limit uint) ([]ExchangeRate, error)

	updateAssetPrice    func(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error)
	upsertAsset         func(ctx context.Context, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error)
	createAssetIfAbsent func(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error)
	countContributed    func(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	getAssets           func(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error)
	getAssetByID        func(ctx context.Context, assetID uuid.UUID) (Asset, error)
	searchAssets        func(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error)
}

func (f *fakeRepository) UpsertExchangeRate(ctx context.Context, from, to money.Currency, rate decimal.Decimal, rateDate time.Time) (ExchangeRate, error) {
	return f.upsertExchangeRate(ctx, from, to, rate, rateDate)
}

func (f *fakeRepository) UpsertPublicExchangeRate(ctx context.Context, from, to money.Currency, rate decimal.Decimal, rateDate time.Time, source ProviderID) (ExchangeRate, error) {
	return f.upsertPublicExchangeRate(ctx, from, to, rate, rateDate, source)
}

func (f *fakeRepository) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	return f.getExchangeRates(ctx, offset, limit)
}

func (f *fakeRepository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	if f.updateAssetPrice == nil {
		return Asset{}, nil
	}
	return f.updateAssetPrice(ctx, assetID, price)
}

func (f *fakeRepository) UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
	if f.upsertAsset == nil {
		return Asset{}, nil
	}
	return f.upsertAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (f *fakeRepository) CreateAssetIfAbsent(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
	if f.createAssetIfAbsent == nil {
		return Asset{}, nil
	}
	return f.createAssetIfAbsent(ctx, userID, ticker, name, assetType, exchange, currency)
}

// CountAssetsContributedBy defaults to an empty quota so the scenarios that do
// not care about it are not made to stub it.
func (f *fakeRepository) CountAssetsContributedBy(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	if f.countContributed == nil {
		return 0, nil
	}
	return f.countContributed(ctx, userID, since)
}

func (f *fakeRepository) GetAssets(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error) {
	if f.getAssets == nil {
		return nil, nil
	}
	return f.getAssets(ctx, view, offset, limit)
}

func (f *fakeRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	if f.getAssetByID == nil {
		return Asset{}, nil
	}
	return f.getAssetByID(ctx, assetID)
}

func (f *fakeRepository) SearchAssets(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error) {
	if f.searchAssets == nil {
		return nil, nil
	}
	return f.searchAssets(ctx, view, search, offset, limit)
}

// fakePriceProvider stubs the market data provider used by the sync jobs.
type fakePriceProvider struct {
	fetchQuote        func(ctx context.Context, symbol string) (marketdata.QuoteResult, error)
	fetchExchangeRate func(ctx context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error)
}

func (p *fakePriceProvider) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	return p.fetchQuote(ctx, symbol)
}

func (p *fakePriceProvider) FetchExchangeRate(ctx context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error) {
	return p.fetchExchangeRate(ctx, from, to)
}

// memStorage is an in-memory fiber.Storage that honours TTLs, good enough to
// exercise the sync-marker caching logic without Redis.
type memStorage struct {
	mu    sync.Mutex
	items map[string]memItem
}

type memItem struct {
	value     []byte
	expiresAt time.Time
}

func newMemStorage() *memStorage {
	return new(memStorage{items: map[string]memItem{}})
}

func (s *memStorage) GetWithContext(_ context.Context, key string) ([]byte, error) {
	return s.Get(key)
}

func (s *memStorage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok {
		return nil, nil
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(s.items, key)
		return nil, nil
	}
	return item.value, nil
}

func (s *memStorage) SetWithContext(_ context.Context, key string, val []byte, exp time.Duration) error {
	return s.Set(key, val, exp)
}

func (s *memStorage) Set(key string, val []byte, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := memItem{value: append([]byte(nil), val...)}
	if exp > 0 {
		item.expiresAt = time.Now().Add(exp)
	}
	s.items[key] = item
	return nil
}

func (s *memStorage) DeleteWithContext(_ context.Context, key string) error {
	return s.Delete(key)
}

func (s *memStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *memStorage) ResetWithContext(_ context.Context) error {
	return s.Reset()
}

func (s *memStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string]memItem{}
	return nil
}

func (s *memStorage) Close() error { return nil }

// mustUSD parses amount as a USD money.Money, failing the test on error.
func mustUSD(t *testing.T, amount string) money.Money {
	t.Helper()

	m, err := money.NewMoneyFromString(amount, money.USD)
	if err != nil {
		t.Fatalf("mustUSD(%q): %v", amount, err)
	}

	return m
}

func newTestServices(repo Repository, storage *memStorage) *service {
	return newService(repo, storage, nil, nil, testKeyring(), logger.Noop())
}

// fakePublicRateSource stands in for the keyless feed. It takes no credential,
// like the real one, so there is nothing to fake but the answer.
type fakePublicRateSource struct {
	rates []marketdata.PublicRate
	err   error
	calls int
}

func (f *fakePublicRateSource) FetchRates(context.Context) ([]marketdata.PublicRate, error) {
	f.calls++

	return f.rates, f.err
}

// fakeFactory stands in for marketdata/providers. It ignores the keys and
// returns a canned provider, so tests exercise the sync logic rather than the
// chain assembly (which providers has its own tests for).
type fakeFactory struct {
	provider marketdata.Provider
	// gotCreds records what the service handed over, so a test can assert the
	// right user's keys were opened.
	gotCreds []marketdata.Credential
	err      error
}

func (f *fakeFactory) For(creds []marketdata.Credential) (marketdata.Provider, error) {
	f.gotCreds = creds

	if f.err != nil {
		return nil, f.err
	}
	if len(creds) == 0 {
		return nil, marketdata.ErrNoCredentials
	}

	return f.provider, nil
}

// testKeyring builds a real keyring over a throwaway key: the sealing path is
// cheap and exercising it for real is worth more than stubbing it.
func testKeyring() *secretbox.Keyring {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("rand: " + err.Error())
	}

	ring, err := secretbox.NewKeyring([]string{"1:" + base64.StdEncoding.EncodeToString(key)}, 1)
	if err != nil {
		panic("keyring: " + err.Error())
	}

	return ring
}

// credentialStore is an in-memory CredentialStore, enough to drive the BYO-key
// sync and the credential use cases without Postgres.
//
// It mirrors the two behaviours of the Postgres implementation that the service
// actually leans on: the sync-facing reads skip keys known to be invalid, and a
// status written against a credential that does not exist is an error rather
// than a silent no-op.
type credentialStore struct {
	mu     sync.Mutex
	sealed map[uuid.UUID][]sealedCredential
	meta   map[string]Credential
	prices map[string]money.Money
	rates  map[string]decimal.Decimal
}

func newCredentialStore() *credentialStore {
	return new(credentialStore{
		sealed: map[uuid.UUID][]sealedCredential{},
		meta:   map[string]Credential{},
		prices: map[string]money.Money{},
		rates:  map[string]decimal.Decimal{},
	})
}

func credKey(userID uuid.UUID, provider ProviderID) string {
	return userID.String() + "/" + string(provider)
}

// seed stores a key for a user, sealed exactly as production would.
func (c *credentialStore) seed(t *testing.T, ring *secretbox.Keyring, userID uuid.UUID, provider ProviderID, apiKey string) {
	t.Helper()

	sealed, err := ring.Seal([]byte(apiKey), credentialAAD(userID.String(), string(provider)))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.sealed[userID] = append(c.sealed[userID], sealedCredential{Provider: provider, Sealed: sealed})
	c.meta[credKey(userID, provider)] = Credential{
		Provider: provider,
		Last4:    last4(apiKey),
		Status:   CredentialActive,
	}
}

// GetSealedCredentials skips keys the provider already rejected, like the
// WHERE status <> 'invalid' of the real query.
func (c *credentialStore) GetSealedCredentials(_ context.Context, userID uuid.UUID) ([]sealedCredential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	usable := make([]sealedCredential, 0, len(c.sealed[userID]))
	for _, sc := range c.sealed[userID] {
		if c.meta[credKey(userID, sc.Provider)].Status != CredentialInvalid {
			usable = append(usable, sc)
		}
	}

	return usable, nil
}

func (c *credentialStore) GetSealedCredential(_ context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sc := range c.sealed[userID] {
		if sc.Provider == provider {
			return sc, nil
		}
	}

	return sealedCredential{}, ErrCredentialNotFound
}

func (c *credentialStore) SetCredentialStatus(_ context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, lastErr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := credKey(userID, provider)

	cred, ok := c.meta[key]
	if !ok {
		return ErrCredentialNotFound
	}

	cred.Status, cred.LastError = status, lastErr
	if status == CredentialActive {
		cred.LastVerifiedAt = new(time.Now().UTC())
	}
	c.meta[key] = cred

	return nil
}

// MarkCredentialWorking mirrors the real UPDATE … WHERE status <> 'active':
// a no-op on a healthy key, and never an error when the row is gone.
func (c *credentialStore) MarkCredentialWorking(_ context.Context, userID uuid.UUID, provider ProviderID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := credKey(userID, provider)

	cred, ok := c.meta[key]
	if !ok || cred.Status == CredentialActive {
		return nil
	}

	cred.Status, cred.LastError = CredentialActive, ""
	cred.LastVerifiedAt = new(time.Now().UTC())
	c.meta[key] = cred

	return nil
}

func (c *credentialStore) statusOf(userID uuid.UUID, provider ProviderID) CredentialStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.meta[credKey(userID, provider)].Status
}

func (c *credentialStore) UpsertUserAssetPrice(_ context.Context, userID, assetID uuid.UUID, price money.Money, _ money.Currency, _ ProviderID, _ time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[userID.String()+"/"+assetID.String()] = price

	return nil
}

func (c *credentialStore) priceOf(userID, assetID uuid.UUID) (money.Money, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.prices[userID.String()+"/"+assetID.String()]

	return p, ok
}

func (c *credentialStore) UpsertUserExchangeRate(_ context.Context, userID uuid.UUID, from, to money.Currency, rate decimal.Decimal, _ ProviderID, _ time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates[userID.String()+"/"+from.String()+to.String()] = rate

	return nil
}

func (c *credentialStore) UsersWithCredentials(context.Context) ([]uuid.UUID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ids := make([]uuid.UUID, 0, len(c.sealed))
	for id := range c.sealed {
		ids = append(ids, id)
	}

	return ids, nil
}

func (c *credentialStore) UpsertCredential(_ context.Context, userID uuid.UUID, cred sealedCredential, keyLast4 string, status CredentialStatus, verifiedAt *time.Time) (Credential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	replaced := false
	for i, sc := range c.sealed[userID] {
		if sc.Provider == cred.Provider {
			c.sealed[userID][i], replaced = cred, true

			break
		}
	}
	if !replaced {
		c.sealed[userID] = append(c.sealed[userID], cred)
	}

	now := time.Now().UTC()
	stored := Credential{
		Provider:       cred.Provider,
		Last4:          keyLast4,
		Status:         status,
		LastVerifiedAt: verifiedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	c.meta[credKey(userID, cred.Provider)] = stored

	return stored, nil
}

func (c *credentialStore) ListCredentials(_ context.Context, userID uuid.UUID) ([]Credential, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	creds := make([]Credential, 0, len(c.sealed[userID]))
	for _, sc := range c.sealed[userID] {
		if cred, ok := c.meta[credKey(userID, sc.Provider)]; ok {
			creds = append(creds, cred)
		}
	}

	slices.SortFunc(creds, func(a, b Credential) int {
		return strings.Compare(string(a.Provider), string(b.Provider))
	})

	return creds, nil
}

func (c *credentialStore) DeleteCredential(_ context.Context, userID uuid.UUID, provider ProviderID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := credKey(userID, provider)
	if _, ok := c.meta[key]; !ok {
		return ErrCredentialNotFound
	}

	delete(c.meta, key)
	c.sealed[userID] = slices.DeleteFunc(c.sealed[userID], func(sc sealedCredential) bool {
		return sc.Provider == provider
	})

	return nil
}

// The CredentialStore half of Repository, forwarded to the in-memory store.
// Explicit forwarding rather than embedding: embedding both Repository and
// CredentialStore would make every one of these selectors ambiguous.

func (f *fakeRepository) UpsertCredential(ctx context.Context, userID uuid.UUID, cred sealedCredential, keyLast4 string, status CredentialStatus, verifiedAt *time.Time) (Credential, error) {
	return f.creds.UpsertCredential(ctx, userID, cred, keyLast4, status, verifiedAt)
}

func (f *fakeRepository) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	return f.creds.ListCredentials(ctx, userID)
}

func (f *fakeRepository) GetSealedCredentials(ctx context.Context, userID uuid.UUID) ([]sealedCredential, error) {
	return f.creds.GetSealedCredentials(ctx, userID)
}

func (f *fakeRepository) GetSealedCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error) {
	return f.creds.GetSealedCredential(ctx, userID, provider)
}

func (f *fakeRepository) DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	return f.creds.DeleteCredential(ctx, userID, provider)
}

func (f *fakeRepository) SetCredentialStatus(ctx context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, lastErr string) error {
	return f.creds.SetCredentialStatus(ctx, userID, provider, status, lastErr)
}

func (f *fakeRepository) MarkCredentialWorking(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	return f.creds.MarkCredentialWorking(ctx, userID, provider)
}

func (f *fakeRepository) UsersWithCredentials(ctx context.Context) ([]uuid.UUID, error) {
	return f.creds.UsersWithCredentials(ctx)
}

func (f *fakeRepository) UpsertUserAssetPrice(ctx context.Context, userID, assetID uuid.UUID, price money.Money, currency money.Currency, source ProviderID, fetchedAt time.Time) error {
	return f.creds.UpsertUserAssetPrice(ctx, userID, assetID, price, currency, source, fetchedAt)
}

func (f *fakeRepository) UpsertUserExchangeRate(ctx context.Context, userID uuid.UUID, from, to money.Currency, rate decimal.Decimal, source ProviderID, fetchedAt time.Time) error {
	return f.creds.UpsertUserExchangeRate(ctx, userID, from, to, rate, source, fetchedAt)
}
