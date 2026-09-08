package portfolio

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"
)

// The consolidated asset list lives entirely in SQL — the grouping, the price
// pick, the conversion and the units — so the fake repository the rest of the
// suite runs on cannot see any of it. It needs the same database the growth
// tests do, and skips the same way without TEST_DATABASE_URL:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/

// heldAcrossPortfolios plants an account holding the same asset in two
// portfolios plus, one at a time, each case the query has to get right: an
// asset quoted in another currency with a rate, one without a rate, one with no
// price at all, and a position sold down to nothing.
func heldAcrossPortfolios(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	userID := uuid.New()
	portfolioA, portfolioB := uuid.New(), uuid.New()

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM exchange_rates WHERE from_currency = 'NOK' AND to_currency = 'USD'`)
	})

	exec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	      VALUES ($1, 'holdings probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")

	sourceID := uuid.New()
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)

	for _, p := range []struct {
		id       uuid.UUID
		name     string
		currency string
	}{{portfolioA, "probe A", "USD"}, {portfolioB, "probe B", "EUR"}} {
		exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
		      VALUES ($1, $2, $3, 'stocks', (SELECT id FROM risks LIMIT 1), $4)`,
			p.id, userID, p.name, p.currency)
	}

	// Neither currency here is one the application converts for real, and that
	// is the point on both sides.
	//
	// exchange_rates is global — one row per pair, shared by every account — so
	// a fixture that writes a rate for a currency the app actually uses edits
	// production-shaped data. This one did: it upserted EUR→USD at 1.20 on every
	// run, overwriting whatever the ECB feed had synced.
	//
	// NOK is outside platform/currency.Supported, so no sync will ever fetch or
	// overwrite it, and the row is deleted again on cleanup. XTS is ISO 4217's
	// code reserved for testing, which is exactly the guarantee the unconvertible
	// case needs: no feed publishes it, so "no rate in any direction" is true by
	// construction rather than by the table happening to be empty. The previous
	// choice, JPY, is in Supported and does have a rate the moment the sync has
	// run once — which is why this assertion failed against any real database.
	//
	// The conflict target is the pair alone: 000014 collapsed the table to one
	// row per pair and re-keyed the unique index, so the three-column target
	// this fixture used to name has not existed since.
	exec(`INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date)
	      VALUES ('NOK', 'USD', 1.20, CURRENT_DATE)
	      ON CONFLICT (from_currency, to_currency) DO UPDATE SET rate = 1.20`)

	asset := func(ticker, currency string, price any) uuid.UUID {
		t.Helper()
		id := uuid.New()
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
		      VALUES ($1, $2, $3, 'stock', $4, $5)`, id, ticker+uuid.New().String()[:6], ticker, currency, price)

		return id
	}

	entry := func(portfolioID, assetID uuid.UUID, quantity, price float64) {
		t.Helper()
		exec(`INSERT INTO portfolio_entries
		        (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, $5, 'USD', $6)`,
			portfolioID, assetID, sourceID, quantity, price, time.Now())
	}

	priced := asset("PRICED", "USD", 200)
	entry(portfolioA, priced, 10, 150)
	entry(portfolioB, priced, 5, 180)

	// Priced with the user's own key, which outranks the catalog's 300. It is
	// also the only case that has a provider to name and an hour to report:
	// a manual catalog price and a cost basis come from nobody.
	owned := asset("OWNED", "USD", 300)
	entry(portfolioA, owned, 2, 250)
	exec(`INSERT INTO user_asset_prices (user_id, asset_id, price, currency, source, fetched_at)
	      VALUES ($1, $2, 400, 'USD', 'finnhub', NOW())`, userID, owned)

	inNOK := asset("INNOK", "NOK", 50)
	entry(portfolioA, inNOK, 4, 40)

	atCost := asset("ATCOST", "USD", nil)
	entry(portfolioA, atCost, 2, 100)

	noRate := asset("NORATE", "XTS", 10)
	entry(portfolioA, noRate, 3, 9)

	soldOut := asset("SOLDOUT", "USD", 500)
	entry(portfolioA, soldOut, 0, 400)

	return userID
}

// worth compares an amount against a round figure on the decimal engine: the
// query returns text at eight decimals, and asserting on that string would be
// asserting on Postgres's formatting rather than on the value.
func worth(t *testing.T, label, amount string, want int) {
	t.Helper()
	if amountOf(amount).Cmp(decimalFromInt(want)) != 0 {
		t.Errorf("%s market value = %q, want %d", label, amount, want)
	}
}

func TestAssetHoldingsTotalTheSameAssetAcrossPortfolios(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossPortfolios(t, pool)

	holdings, err := repo.GetAssetHoldingsByUserID(context.Background(), userID, money.USD)
	if err != nil {
		t.Fatalf("GetAssetHoldingsByUserID: %v", err)
	}

	byName := make(map[string]AssetHolding, len(holdings))
	for _, h := range holdings {
		byName[h.Name] = h
	}

	// A position sold in full is not something the user holds; it must not
	// reach the list as a zero-unit row.
	if _, listed := byName["SOLDOUT"]; listed {
		t.Error("a fully sold position is listed as a holding")
	}
	if len(holdings) != 5 {
		t.Fatalf("holdings = %d, want 5: %+v", len(holdings), holdings)
	}

	// The whole point of the view: 10 units in one portfolio and 5 in another
	// are one row of 15, valued at the asset's price.
	priced := byName["PRICED"]
	if amountOf(priced.Quantity).Cmp(decimalFromInt(15)) != 0 {
		t.Errorf("PRICED quantity = %q, want 15 (10 in one portfolio, 5 in the other)", priced.Quantity)
	}
	if priced.Portfolios != 2 {
		t.Errorf("PRICED portfolios = %d, want 2", priced.Portfolios)
	}
	// 15 × 200.
	worth(t, "PRICED", priced.MarketValue, 3000)
	if priced.PriceSource != PriceSourceManual {
		t.Errorf("PRICED price source = %q, want manual", priced.PriceSource)
	}

	// The user's own price wins over the catalog's, and it is the one case that
	// says where it came from. Without the provider and the timestamp,
	// priceSource 'own' tells a client the price is theirs and leaves it with no
	// way to decide whether to re-ask, or whom.
	owned := byName["OWNED"]
	if owned.PriceSource != PriceSourceOwn {
		t.Errorf("OWNED price source = %q, want own", owned.PriceSource)
	}
	// 2 × 400, the user's price, not the catalog's 300.
	worth(t, "OWNED", owned.MarketValue, 800)
	if owned.PriceProvider != "finnhub" {
		t.Errorf("OWNED price provider = %q, want finnhub", owned.PriceProvider)
	}
	if owned.PriceFetchedAt == nil {
		t.Error("OWNED has no price timestamp")
	}

	// The mirror image, and the reason both columns are NULL-guarded in the
	// query rather than just selected: a price nobody fetched has nobody to
	// attribute and no hour to report.
	if priced.PriceProvider != "" || priced.PriceFetchedAt != nil {
		t.Errorf("PRICED is a manual catalog price but claims %q at %v", priced.PriceProvider, priced.PriceFetchedAt)
	}

	// Priced in another currency, reported in dollars: the price stays in the
	// asset's own currency and only the total is converted.
	inNOK := byName["INNOK"]
	// 4 × 50 × 1.20.
	worth(t, "INNOK", inNOK.MarketValue, 240)
	if inNOK.Currency != money.NOK {
		t.Errorf("INNOK currency = %v, want NOK (the asset's own)", inNOK.Currency)
	}
	if inNOK.DisplayCurrency != money.USD {
		t.Errorf("INNOK display currency = %v, want USD", inNOK.DisplayCurrency)
	}
	if inNOK.PositionsUnconverted != 0 {
		t.Errorf("INNOK unconverted = %d, want 0: the rate exists", inNOK.PositionsUnconverted)
	}

	// No price anywhere: the position is carried at what it cost, and no single
	// number stands for the asset's price.
	atCost := byName["ATCOST"]
	if atCost.PriceSource != PriceSourceCost {
		t.Errorf("ATCOST price source = %q, want cost", atCost.PriceSource)
	}
	if atCost.PriceProvider != "" || atCost.PriceFetchedAt != nil {
		t.Errorf("ATCOST is carried at cost but claims %q at %v", atCost.PriceProvider, atCost.PriceFetchedAt)
	}
	if atCost.MarketPrice != "" {
		t.Errorf("ATCOST market price = %q, want empty", atCost.MarketPrice)
	}
	// 2 × 100, its cost.
	worth(t, "ATCOST", atCost.MarketValue, 200)

	// No rate in any direction: counted at face value and flagged, the same
	// choice the summary and the allocation make.
	noRate := byName["NORATE"]
	if noRate.PositionsUnconverted != 1 {
		t.Errorf("NORATE unconverted = %d, want 1", noRate.PositionsUnconverted)
	}
	// 3 × 10, unconverted.
	worth(t, "NORATE", noRate.MarketValue, 30)

	// Ordered by what they are worth, so the list opens on what matters most
	// and the pie chart's first slices are its first rows.
	for i := 1; i < len(holdings); i++ {
		if amountOf(holdings[i-1].MarketValue).Cmp(amountOf(holdings[i].MarketValue)) < 0 {
			t.Errorf("holding %d (%s) is worth less than the one after it", i-1, holdings[i-1].Name)
		}
	}
}

// Omitting the currency means "the account's preferred one", the same contract
// the summary and the allocation keep. Only the SQL can prove it: money.XXX
// reaching the query as the literal "XXX" survives the NULLIF and becomes the
// target currency, converting nothing and labelling everything ¤.
func TestAssetHoldingsWithoutACurrencyUseTheAccountPreferredOne(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossPortfolios(t, pool)

	holdings, err := repo.GetAssetHoldingsByUserID(context.Background(), userID, money.XXX)
	if err != nil {
		t.Fatalf("GetAssetHoldingsByUserID: %v", err)
	}
	if len(holdings) == 0 {
		t.Fatal("the list came back empty")
	}

	for _, h := range holdings {
		if h.DisplayCurrency != money.USD {
			t.Fatalf("%s came back in %v, want USD (the account's preferred currency)", h.Name, h.DisplayCurrency)
		}
	}
}
