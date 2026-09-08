package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// GetAssetHoldingsByUserID totals what the user holds of each asset across
// every portfolio they own.
//
// It is the per-asset sibling of GetAssetAllocationByUserID and shares its
// valuation rules on purpose, so the list and the donut cannot disagree about
// the same position:
//
//   - the price is the user's own (user_asset_prices), then the operator's
//     manual one (assets.current_price), then the entry's cost;
//   - the currency that price is quoted in is chosen with it — a position
//     carried at cost is in its cost currency, not in the asset's;
//   - conversion goes through fx_rate, and a position with no rate is still
//     counted at face value and reported in PositionsUnconverted.
//
// The two things it adds are the ones only a consolidated view can answer:
// Quantity, the units held regardless of portfolio, and Portfolios, how many of
// them the asset is spread over.
//
// Entries at quantity zero are left out, matching the allocation: a position
// sold in full is not something the user holds.
func (r *PostgresRepository) GetAssetHoldingsByUserID(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]AssetHolding, error) {
	// Same rounding as the allocation, for the same reason: Postgres adds the
	// scales of what it multiplies, and past nineteen decimals the engine that
	// parses this text back reads zero.
	rows, err := r.db.Query(ctx, `
		SELECT
			a.id,
			a.ticker,
			a.name,
			a.asset_type,
			COALESCE(a.exchange, ''),
			a.currency,
			SUM(pe.quantity::numeric)::text AS quantity,
			-- The price the asset itself carries, if any. uap and assets are
			-- both keyed by the asset (uap also by the user, fixed by the WHERE
			-- below), so it is one value per group; the aggregate is what tells
			-- Postgres that. NULL means every entry fell through to its own
			-- cost, and those differ from one another — hence the empty string
			-- rather than a number that stands for none of them.
			COALESCE(COALESCE(MAX(uap.price), MAX(a.current_price))::text, '') AS market_price,
			ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8)::text AS market_value,
			target.code,
			COUNT(DISTINCT pe.portfolio_id)::bigint AS portfolios,
			CASE
				WHEN MAX(uap.price)       IS NOT NULL THEN 'own'
				WHEN MAX(a.current_price) IS NOT NULL THEN 'manual'
				ELSE 'cost'
			END AS price_source,
			-- Who produced that price and when. Aggregated for the same reason
			-- the price above is: uap is one row per (user, asset), so MAX is
			-- how the group is told there is only one value. Both stay NULL
			-- unless the price came from the user's own key — a manual catalog
			-- price and a cost basis have no provider to name.
			CASE WHEN MAX(uap.price) IS NOT NULL THEN MAX(uap.source) END AS price_provider,
			CASE WHEN MAX(uap.price) IS NOT NULL THEN MAX(uap.fetched_at) END AS price_fetched_at,
			COUNT(*) FILTER (WHERE fx.rate IS NULL)::bigint AS positions_unconverted
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN users u      ON u.id = p.user_id
		JOIN assets a     ON a.id = pe.asset_id
		LEFT JOIN user_asset_prices uap ON uap.asset_id = a.id AND uap.user_id = p.user_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(NULLIF($2::text, ''), u.preferred_currency) AS code
		) target
		CROSS JOIN LATERAL (
			SELECT
				COALESCE(uap.price::numeric, a.current_price::numeric, pe.price::numeric) AS price,
				CASE
					WHEN COALESCE(uap.price, a.current_price) IS NOT NULL
						THEN COALESCE(a.currency, pe.cost_currency)
					ELSE pe.cost_currency
				END AS currency
		) v
		CROSS JOIN LATERAL (
			SELECT fx_rate(p.user_id, v.currency, target.code) AS rate
		) fx
		WHERE p.user_id = $1
		  AND pe.quantity::numeric > 0
		GROUP BY a.id, a.ticker, a.name, a.asset_type, a.exchange, a.currency, target.code
		ORDER BY ROUND(COALESCE(SUM(pe.quantity::numeric * v.price * COALESCE(fx.rate, 1)), 0), 8) DESC, a.ticker
	`, userID, currencyParam(targetCurrency))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holdings := make([]AssetHolding, 0)
	for rows.Next() {
		var holding AssetHolding
		// NULL for every row whose price did not come from the user's own key,
		// which is most of them on an account with no provider configured.
		var priceProvider *string

		if err := rows.Scan(
			&holding.AssetID,
			&holding.Ticker,
			&holding.Name,
			&holding.AssetType,
			&holding.Exchange,
			&holding.Currency,
			&holding.Quantity,
			&holding.MarketPrice,
			&holding.MarketValue,
			&holding.DisplayCurrency,
			&holding.Portfolios,
			&holding.PriceSource,
			&priceProvider,
			&holding.PriceFetchedAt,
			&holding.PositionsUnconverted,
		); err != nil {
			return nil, err
		}

		if priceProvider != nil {
			holding.PriceProvider = *priceProvider
		}

		holdings = append(holdings, holding)
	}

	return holdings, rows.Err()
}
