package market

import (
	"context"
	"errors"
	"strings"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/database"
)

// assetColumns is the projection every catalog read shares, aliased to a so the
// visibility join can be appended to any of them.
const assetColumns = `a.id, a.ticker, a.name, a.asset_type, COALESCE(a.exchange, ''), a.currency,
	a.current_price, a.price_updated_at, a.is_curated, a.created_at, a.updated_at`

// visibleAssets is the FROM clause of a catalog read: the assets table with the
// viewer's membership attached. $1 is the "see everything" flag (admins), $2
// the viewer's id.
//
// Curated rows are visible to everybody; a contributed row is visible only to
// the users in user_catalog_assets. Written as a LEFT JOIN rather than an
// EXISTS subquery so the ordering below can also read from it.
const visibleAssets = `
	FROM assets a
	LEFT JOIN user_catalog_assets uca ON uca.asset_id = a.id AND uca.user_id = $2
	WHERE ($1::boolean OR a.is_curated OR uca.user_id IS NOT NULL)`

// scanAssets reads a result set produced with assetColumns.
func scanAssets(rows pgx.Rows) ([]Asset, error) {
	defer rows.Close()

	assets := make([]Asset, 0)

	for rows.Next() {
		var asset Asset

		if err := rows.Scan(
			&asset.ID,
			&asset.Ticker,
			&asset.Name,
			&asset.AssetType,
			&asset.Exchange,
			&asset.Currency,
			&asset.CurrentPrice,
			&asset.PriceUpdatedAt,
			&asset.IsCurated,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if asset.CurrentPrice != nil {
			asset.CurrentPrice.SetCurrency(asset.Currency)
		}

		assets = append(assets, asset)
	}

	return assets, rows.Err()
}

// GetAssetByID is deliberately not scoped to a viewer: it answers reads that
// already start from an id the caller is entitled to — the sync walking a
// user's own holdings, an entry resolving the asset it points at — and scoping
// it would break a portfolio whose asset was contributed by somebody else.
func (r *PostgresRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	var asset Asset

	err := r.db.QueryRow(ctx, `
		SELECT `+assetColumns+`
		FROM assets a WHERE a.id = $1
	`, assetID).Scan(
		&asset.ID, &asset.Ticker, &asset.Name, &asset.AssetType, &asset.Exchange,
		&asset.Currency, &asset.CurrentPrice, &asset.PriceUpdatedAt, &asset.IsCurated,
		&asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, ErrAssetNotFound
		}

		return Asset{}, err
	}

	if asset.CurrentPrice != nil {
		asset.CurrentPrice.SetCurrency(asset.Currency)
	}

	return asset, nil
}

func (r *PostgresRepository) GetAssets(ctx context.Context, view CatalogView, offset, limit uint) ([]Asset, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+assetColumns+visibleAssets+`
		ORDER BY a.is_curated DESC, a.ticker
		LIMIT $3 OFFSET $4
	`, view.All, view.ViewerID, limit, offset)
	if err != nil {
		return []Asset{}, err
	}

	return scanAssets(rows)
}

func (r *PostgresRepository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error) {
	var asset Asset

	err := r.db.QueryRow(ctx, `
		UPDATE assets a
		SET current_price = $2::numeric, price_updated_at = NOW()
		WHERE a.id = $1
		RETURNING `+assetColumns+`
	`, assetID, price.String()).Scan(
		&asset.ID,
		&asset.Ticker,
		&asset.Name,
		&asset.AssetType,
		&asset.Exchange,
		&asset.Currency,
		&asset.CurrentPrice,
		&asset.PriceUpdatedAt,
		&asset.IsCurated,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, ErrAssetNotFound
		}
		return Asset{}, err
	}

	if asset.CurrentPrice != nil {
		asset.CurrentPrice.SetCurrency(asset.Currency)
	}

	return asset, nil
}

// pgUniqueViolation is the SQLSTATE Postgres returns for a unique constraint
// violation (23505). On this table it can only be idx_assets_ticker_exchange:
// the edit renamed a row onto a (ticker, exchange) pair that already exists.
const pgUniqueViolation = "23505"

// UpdateAsset writes an operator's edit of a catalog row.
//
// The price is the only field that is not a plain assignment. A nil Price
// normally leaves the stored number and its timestamp alone — but the currency
// is what gives that number its meaning, so when the edit re-denominates the
// asset without supplying a new price, both are cleared. Keeping them would
// leave the catalog showing a figure in a currency nobody quoted it in, and
// showing nothing is the truthful answer until an admin or a sync fills it
// again. a.currency in the CASE is the row's old value: in an UPDATE every SET
// expression reads the pre-update row.
func (r *PostgresRepository) UpdateAsset(ctx context.Context, assetID uuid.UUID, upd AssetUpdate) (Asset, error) {
	var asset Asset

	// Both optional fields reach the query as an untyped nil when unset, which
	// pgx sends as NULL for the casts below to fall back on.
	var curated, price any
	if upd.IsCurated != nil {
		curated = *upd.IsCurated
	}
	if upd.Price != nil {
		price = upd.Price.String()
	}

	err := r.db.QueryRow(ctx, `
		UPDATE assets a
		SET ticker = $2,
		    name = $3,
		    asset_type = $4::asset_type,
		    exchange = NULLIF($5, ''),
		    currency = $6,
		    is_curated = COALESCE($7::boolean, a.is_curated),
		    current_price = CASE
		        WHEN $8::numeric IS NOT NULL THEN $8::numeric
		        WHEN a.currency IS DISTINCT FROM $6 THEN NULL
		        ELSE a.current_price END,
		    price_updated_at = CASE
		        WHEN $8::numeric IS NOT NULL THEN NOW()
		        WHEN a.currency IS DISTINCT FROM $6 THEN NULL
		        ELSE a.price_updated_at END,
		    updated_at = NOW()
		WHERE a.id = $1
		RETURNING `+assetColumns+`
	`, assetID, upd.Ticker, upd.Name, upd.AssetType, upd.Exchange, upd.Currency, curated, price).Scan(
		&asset.ID,
		&asset.Ticker,
		&asset.Name,
		&asset.AssetType,
		&asset.Exchange,
		&asset.Currency,
		&asset.CurrentPrice,
		&asset.PriceUpdatedAt,
		&asset.IsCurated,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, ErrAssetNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return Asset{}, errAssetDuplicate
		}

		return Asset{}, err
	}

	if asset.CurrentPrice != nil {
		asset.CurrentPrice.SetCurrency(asset.Currency)
	}

	return asset, nil
}

func (r *PostgresRepository) SearchAssets(ctx context.Context, view CatalogView, search string, offset, limit uint) ([]Asset, error) {
	pattern := "%" + strings.ToUpper(strings.TrimSpace(search)) + "%"
	rows, err := r.db.Query(ctx, `
		SELECT `+assetColumns+visibleAssets+`
		  AND (UPPER(a.ticker) LIKE $3 OR UPPER(a.name) LIKE $3)
		ORDER BY a.is_curated DESC, a.ticker
		LIMIT $4 OFFSET $5
	`, view.All, view.ViewerID, pattern, limit, offset)
	if err != nil {
		return []Asset{}, err
	}

	return scanAssets(rows)
}

func (r *PostgresRepository) UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
	var asset Asset

	err := r.db.QueryRow(ctx, `
		INSERT INTO assets (ticker, name, asset_type, exchange, currency, is_curated, created_at, updated_at)
		VALUES ($1, $2, $3::asset_type, NULLIF($4, ''), $5, TRUE, NOW(), NOW())
		ON CONFLICT (ticker, COALESCE(exchange, ''))
		DO UPDATE SET name = EXCLUDED.name, asset_type = EXCLUDED.asset_type, currency = EXCLUDED.currency,
		              is_curated = TRUE, updated_at = NOW()
		RETURNING id, ticker, name, asset_type, COALESCE(exchange, ''), currency, current_price, price_updated_at, is_curated, created_at, updated_at
	`, ticker, name, assetType, exchange, currency).Scan(
		&asset.ID,
		&asset.Ticker,
		&asset.Name,
		&asset.AssetType,
		&asset.Exchange,
		&asset.Currency,
		&asset.CurrentPrice,
		&asset.PriceUpdatedAt,
		&asset.IsCurated,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		return asset, err
	}

	if asset.CurrentPrice != nil {
		asset.CurrentPrice.SetCurrency(asset.Currency)
	}

	return asset, nil
}

// CreateAssetIfAbsent resolves a ticker to a catalog row for one user and makes
// sure that row is in their catalog, in a single transaction.
//
// Two things it deliberately does not do. It never updates an existing row:
// the conflict target is the ticker, so an UPDATE here would let any user
// rewrite the name, type or currency of an asset every other user is holding —
// which is exactly what UpsertAsset does, and why that one stays admin-only.
// And it matches on the ticker alone, ignoring the exchange, for the same
// reason the transaction importer does: a user typing AAPL means the AAPL
// already in the catalog, not a second row that would split the holdings.
//
// The membership row is written even when the asset already existed, so the
// second user to reach for a contributed ticker can find it afterwards.
func (r *PostgresRepository) CreateAssetIfAbsent(ctx context.Context, userID uuid.UUID, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
	var asset Asset

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var assetID uuid.UUID

		err := tx.QueryRow(ctx, `
		SELECT id FROM assets WHERE UPPER(ticker) = $1 ORDER BY created_at LIMIT 1
	`, ticker).Scan(&assetID)

		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `
			INSERT INTO assets (ticker, name, asset_type, exchange, currency, created_by, is_curated, created_at, updated_at)
			VALUES ($1, $2, $3::asset_type, NULLIF($4, ''), $5, $6, FALSE, NOW(), NOW())
			ON CONFLICT (ticker, COALESCE(exchange, '')) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`, ticker, name, assetType, exchange, currency, userID).Scan(&assetID)
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
		INSERT INTO user_catalog_assets (user_id, asset_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, assetID); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, `
		SELECT `+assetColumns+`
		FROM assets a WHERE a.id = $1
	`, assetID).Scan(
			&asset.ID, &asset.Ticker, &asset.Name, &asset.AssetType, &asset.Exchange,
			&asset.Currency, &asset.CurrentPrice, &asset.PriceUpdatedAt, &asset.IsCurated,
			&asset.CreatedAt, &asset.UpdatedAt,
		); err != nil {
			return err
		}

		if asset.CurrentPrice != nil {
			asset.CurrentPrice.SetCurrency(asset.Currency)
		}

		return nil
	}); err != nil {
		return asset, err
	}

	return asset, nil
}

// CountAssetsContributedBy counts the rows this user actually created since a
// point in time. Memberships are not counted: attaching to an asset somebody
// else already contributed adds nothing to the catalog, so it should not spend
// the quota that exists to bound how much the catalog can grow.
func (r *PostgresRepository) CountAssetsContributedBy(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM assets WHERE created_by = $1 AND created_at >= $2
	`, userID, since).Scan(&count)

	return count, err
}
