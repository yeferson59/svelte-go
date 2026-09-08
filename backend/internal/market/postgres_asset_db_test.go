package market

import (
	"context"
	"errors"
	"os"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"
)

// What an operator's edit does to a catalog row lives entirely in SQL: which
// fields fall back to the ones already stored, and what happens to the manual
// price when the currency underneath it changes. The fake repository the rest of
// this package's tests run on cannot see any of it, so these need the same
// database the other *_db_test.go files do, and skip the same way without
// TEST_DATABASE_URL:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/market/

func assetTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omite la prueba contra Postgres")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// plantAsset inserts one curated row priced in USD, which is the state every
// case below edits away from.
func plantAsset(t *testing.T, pool *pgxpool.Pool, ticker, exchange string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	id := uuid.New()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE id = $1`, id)
	})

	if _, err := pool.Exec(ctx, `
		INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, is_curated)
		VALUES ($1, $2, 'Probe Inc.', 'stock', NULLIF($3, ''), 'USD', 190.50, NOW() - INTERVAL '3 days', TRUE)
	`, id, ticker, exchange); err != nil {
		t.Fatalf("plant asset: %v", err)
	}

	return id
}

func TestPostgresUpdateAsset(t *testing.T) {
	pool := assetTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	// A ticker per subtest: the unique index is on (ticker, exchange), and the
	// duplicate case below needs the collision to be the one it planted.
	base := func(ticker string) AssetUpdate {
		return AssetUpdate{
			Ticker:    ticker,
			Name:      "Probe Inc.",
			AssetType: Stock,
			Exchange:  "NASDAQ",
			Currency:  money.USD,
		}
	}

	price := func(t *testing.T, value string, currency money.Currency) *money.Money {
		t.Helper()
		m := mustMoney(t, value, currency)

		return &m
	}

	t.Run("an edit that leaves the currency alone keeps the stored price", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB1", "NASDAQ")

		upd := base("PRB1")
		upd.Name = "Probe Renamed"

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.Name != "Probe Renamed" {
			t.Errorf("name = %q, want the edited one", asset.Name)
		}
		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "190.5" {
			t.Errorf("price = %v, want the stored 190.50 untouched", asset.CurrentPrice)
		}
		// Untouched means untouched: the table is ordered by how old the price
		// is, so bumping the timestamp on a name change would hide a stale one.
		if asset.PriceUpdatedAt == nil {
			t.Fatal("priceUpdatedAt = nil, want the original timestamp")
		}
		if hours := asset.UpdatedAt.Sub(*asset.PriceUpdatedAt).Hours(); hours < 71 {
			t.Errorf("priceUpdatedAt is %.0fh before updatedAt, want the original ~72h", hours)
		}
	})

	t.Run("a new price is written with today's timestamp", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB2", "NASDAQ")

		upd := base("PRB2")
		upd.Price = price(t, "205.25", money.USD)

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "205.25" {
			t.Errorf("price = %v, want 205.25", asset.CurrentPrice)
		}
		if asset.PriceUpdatedAt == nil || asset.UpdatedAt.Sub(*asset.PriceUpdatedAt).Hours() > 1 {
			t.Errorf("priceUpdatedAt = %v, want it sealed now", asset.PriceUpdatedAt)
		}
	})

	t.Run("re-denominating without a price drops the one that no longer means anything", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB3", "NASDAQ")

		upd := base("PRB3")
		upd.Currency = money.COP

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.Currency != money.COP {
			t.Errorf("currency = %s, want COP", asset.Currency)
		}
		if asset.CurrentPrice != nil || asset.PriceUpdatedAt != nil {
			t.Errorf("price = %v (%v), want both cleared: 190.50 USD is not 190.50 COP",
				asset.CurrentPrice, asset.PriceUpdatedAt)
		}
	})

	t.Run("re-denominating with a price keeps the one that was sent", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB4", "NASDAQ")

		upd := base("PRB4")
		upd.Currency = money.COP
		upd.Price = price(t, "812340.75", money.COP)

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "812340.75" {
			t.Errorf("price = %v, want the new one in COP", asset.CurrentPrice)
		}
		if asset.CurrentPrice.GetCurrency() != money.COP {
			t.Errorf("price currency = %s, want COP", asset.CurrentPrice.GetCurrency())
		}
	})

	t.Run("the audience only changes when the edit says so", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB5", "NASDAQ")

		kept, err := repo.UpdateAsset(ctx, id, base("PRB5"))
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if !kept.IsCurated {
			t.Error("isCurated = false after an edit that did not mention it")
		}

		upd := base("PRB5")
		hidden := false
		upd.IsCurated = &hidden

		unpublished, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if unpublished.IsCurated {
			t.Error("isCurated = true, want the row off the shared catalog")
		}
	})

	t.Run("an empty exchange is stored as NULL, so the unique index still pairs", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB6", "NASDAQ")

		upd := base("PRB6")
		upd.Exchange = ""

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if asset.Exchange != "" {
			t.Errorf("exchange = %q, want it cleared", asset.Exchange)
		}
	})

	t.Run("a rename onto an existing pair is a conflict, not a driver error", func(t *testing.T) {
		taken := plantAsset(t, pool, "PRB7", "NASDAQ")
		id := plantAsset(t, pool, "PRB8", "NASDAQ")

		if taken == id {
			t.Fatal("the fixture planted one row for two")
		}

		_, err := repo.UpdateAsset(ctx, id, base("PRB7"))
		if !errors.Is(err, errAssetDuplicate) {
			t.Fatalf("err = %v, want errAssetDuplicate", err)
		}
	})

	t.Run("an asset that is gone answers not found", func(t *testing.T) {
		_, err := repo.UpdateAsset(ctx, uuid.New(), base("PRB9"))
		if !errors.Is(err, ErrAssetNotFound) {
			t.Fatalf("err = %v, want ErrAssetNotFound", err)
		}
	})
}
