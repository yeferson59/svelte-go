// Package portfolio is the investments domain module: portfolios, entries,
// transactions, investment platforms, snapshots and bulk import/export. The
// asset catalog (the Asset type and its lifecycle) is owned by the market
// module; this module references market.Asset and reads asset data through the
// interfaces it declares.
package portfolio

import (
	"encoding/json"
	"fmt"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/market"
)

type SourceType string

const (
	Broker          SourceType = "broker"
	Bank            SourceType = "investment_bank"
	TradingPlatform SourceType = "trading_platform"
	Neobank         SourceType = "neobank"
	DeFi            SourceType = "de_fi"
	CryptoWallet    SourceType = "crypto_wallet"
	MutualFunds     SourceType = "mutual_funds"
	BrokerageHouse  SourceType = "brokerage_house"
	OtherType       SourceType = "other"
)

func (s SourceType) IsValid() bool {
	switch s {
	case Broker, Bank, TradingPlatform, Neobank, DeFi, CryptoWallet, MutualFunds, BrokerageHouse, OtherType:
		return true
	default:
		return false
	}
}

type TransactionType string

const (
	Buy         TransactionType = "buy"
	Sell        TransactionType = "sell"
	Dividend    TransactionType = "dividend"
	Split       TransactionType = "split"
	TransferIn  TransactionType = "transfer_in"
	TransferOut TransactionType = "transfer_out"
	Fee         TransactionType = "fee"
	Interest    TransactionType = "interest"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case Buy, Sell, Dividend, Split, TransferIn, TransferOut, Fee, Interest:
		return true
	default:
		return false
	}
}

type Type string

const (
	TypeStocks                 Type = "stocks"
	TypeETFs                   Type = "etfs"
	TypeCryptos                Type = "cryptos"
	TypeBonds                  Type = "bonds"
	TypeCash                   Type = "cash"
	TypeForex                  Type = "forex"
	TypeRealEstates            Type = "real_estates"
	TypeCommodities            Type = "commodities"
	TypeForexStocks            Type = "forex_stocks"
	TypeForexETFs              Type = "forex_etfs"
	TypeForexCryptos           Type = "forex_cryptos"
	TypeForexBonds             Type = "forex_bonds"
	TypeForexCash              Type = "forex_cash"
	TypeForexRealStates        Type = "forex_real_states"
	TypeForexCommodities       Type = "forex_commodities"
	TypeStocksETFs             Type = "stocks_etfs"
	TypeStocksCryptos          Type = "stocks_cryptos"
	TypeStocksBonds            Type = "stocks_bonds"
	TypeStocksCash             Type = "stocks_cash"
	TypeStocksRealEstates      Type = "stocks_real_estates"
	TypeStocksCommodities      Type = "stocks_commodities"
	TypeETFsCryptos            Type = "etfs_cryptos"
	TypeETFsBonds              Type = "etfs_bonds"
	TypeETFsCash               Type = "etfs_cash"
	TypeETFsRealEstates        Type = "etfs_real_estates"
	TypeETFsCommodities        Type = "etfs_commodities"
	TypeCryptosBonds           Type = "cryptos_bonds"
	TypeCryptosCash            Type = "cryptos_cash"
	TypeCryptosRealEstates     Type = "cryptos_real_estates"
	TypeCryptosCommodities     Type = "cryptos_commodities"
	TypeBondsCash              Type = "bonds_cash"
	TypeBondsRealEstates       Type = "bonds_real_estates"
	TypeBondsCommodities       Type = "bonds_commodities"
	TypeCashRealEstates        Type = "cash_real_estates"
	TypeCashCommodities        Type = "cash_commodities"
	TypeRealEstatesCommodities Type = "real_estates_commodities"
	TypeDiversified            Type = "diversified"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeStocks, TypeETFs, TypeCryptos, TypeBonds,
		TypeCash, TypeForex, TypeRealEstates, TypeCommodities,
		TypeForexStocks, TypeForexETFs, TypeForexCryptos, TypeForexBonds,
		TypeForexCash, TypeForexRealStates, TypeForexCommodities,
		TypeStocksETFs, TypeStocksCryptos, TypeStocksBonds, TypeStocksCash,
		TypeStocksRealEstates, TypeStocksCommodities,
		TypeETFsCryptos, TypeETFsBonds, TypeETFsCash, TypeETFsRealEstates,
		TypeETFsCommodities,
		TypeCryptosBonds, TypeCryptosCash, TypeCryptosRealEstates, TypeCryptosCommodities,
		TypeBondsCash, TypeBondsRealEstates, TypeBondsCommodities,
		TypeCashRealEstates, TypeCashCommodities,
		TypeRealEstatesCommodities, TypeDiversified:
		return true
	default:
		return false
	}
}

type PriceSource string

const (
	// PriceSourceOwn is a price fetched with this user's own provider key. It
	// is the only one that makes the valuation a market valuation.
	PriceSourceOwn PriceSource = "own"
	// PriceSourceManual is the operator's manually entered reference price
	// (assets.current_price, PATCH /portfolios/assets/:id/price). It carries no
	// provider licence and no freshness guarantee.
	PriceSourceManual PriceSource = "manual"
	// PriceSourceCost means no price was available and the position is valued
	// at what it cost. Its gain/loss is zero by construction, so a client must
	// not present it as a return.
	PriceSourceCost PriceSource = "cost"
)

type InvestmentSource struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"userId"`
	Name        string        `json:"name"`
	SourceType  SourceType    `json:"sourceType"`
	Description string        `json:"description"`
	IsActive    bool          `json:"isActive"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	User        identity.User `json:"user,omitzero"`
	Entries     []Entry       `json:"portfolioEntries,omitempty"`
}

type Risk struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Portfolios  []Portfolio `json:"portfolios,omitempty"`
}

type Portfolio struct {
	ID           uuid.UUID      `json:"id"`
	UserID       uuid.UUID      `json:"userId"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Type         Type           `json:"type"`
	RiskID       uuid.UUID      `json:"riskId"`
	BaseCurrency money.Currency `json:"baseCurrency"`
	IsDefault    bool           `json:"isDefault"`
	PriceValue   *money.Money   `json:"priceValue"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	Risk         Risk           `json:"risk,omitzero"`
	User         identity.User  `json:"user,omitzero"`
	Entries      []Entry        `json:"portfolioEntries,omitempty"`
	Snapshots    []Snapshot     `json:"portfolioSnapshots,omitempty"`
}

type Entry struct {
	ID              uuid.UUID        `json:"id"`
	PortfolioID     uuid.UUID        `json:"portfolioId"`
	AssetID         uuid.UUID        `json:"assetId"`
	SourceID        uuid.UUID        `json:"sourceId"`
	Quantity        decimal.Decimal  `json:"quantity"`
	Price           money.Money      `json:"price"`
	CostCurrency    money.Currency   `json:"costCurrency"`
	Category        market.AssetType `json:"category"`
	EntryDate       time.Time        `json:"entryDate"`
	Notes           string           `json:"notes"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	Portfolio       Portfolio        `json:"portfolio,omitzero"`
	Asset           market.Asset     `json:"asset,omitzero"`
	Source          InvestmentSource `json:"source,omitzero"`
	Transactions    []Transaction    `json:"transactions,omitempty"`
	PriceSource     PriceSource      `json:"priceSource,omitempty"`
	CostBasisBase   money.Money      `json:"costBasisBase,omitzero"`
	MarketValueBase money.Money      `json:"marketValueBase,omitzero"`
	FXConverted     bool             `json:"fxConverted,omitempty"`
}

type Transaction struct {
	ID       uuid.UUID       `json:"id"`
	EntryID  uuid.UUID       `json:"entryId"`
	Type     TransactionType `json:"type"`
	Quantity decimal.Decimal `json:"quantity"`
	// Price is in Currency: the currency the trade was quoted in, which is the
	// asset's own whenever the broker fills on its home exchange.
	Price    money.Money    `json:"price"`
	Currency money.Currency `json:"currency"`
	// FXRate is how much of the position's cost currency one unit of Currency
	// bought on the day of the trade, so Price * FXRate is what the trade
	// actually cost. It is 1 whenever the two currencies agree, which is every
	// transaction recorded before 000029.
	FXRate decimal.Decimal `json:"fxRate"`
	// CostCurrency is the position's, not the transaction's — it comes from the
	// entry and is repeated here so a transaction read on its own says what
	// FXRate converts into. Empty on reads that do not join the entry.
	CostCurrency money.Currency `json:"costCurrency,omitzero"`
	// FeesCurrency is Currency or CostCurrency, and says which. A commission
	// rides the trade's conversion only in the first case: brokers bill it to
	// the account as often as to the fill, and the confirmation behind this
	// field quotes a EUR price beside a USD commission.
	FeesCurrency    money.Currency `json:"feesCurrency,omitzero"`
	Fees            money.Money    `json:"fees"`
	TransactionDate time.Time      `json:"transactionDate"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	Entry           Entry          `json:"entry,omitzero"`
}

// TransactionInput is one transaction as a caller states it, on its way to
// being persisted.
//
// It replaced eight positional parameters on four call paths, but the reason it
// is a type rather than a longer signature is Currency and FXRate: those two
// are only meaningful together, and the rule binding them needs a third value
// — the cost currency of the position — that no caller passes explicitly.
// Validate is where that rule lives, so every writer applies the same one.
type TransactionInput struct {
	Type            TransactionType
	Quantity        decimal.Decimal
	Price           money.Money
	Currency        money.Currency
	FXRate          decimal.Decimal
	Fees            money.Money
	FeesCurrency    money.Currency
	TransactionDate time.Time
	Notes           string
}

// Rate resolves the omitted case. A caller that sends no rate at all is the
// common one — same currency on both sides — and 1 is the honest answer there
// rather than a default standing in for a missing fact. Validate is what makes
// sure it is only reached in that case.
func (in TransactionInput) Rate() decimal.Decimal {
	if in.FXRate.IsZero() {
		return decimal.One
	}

	return in.FXRate
}

// Validate checks the input against the cost currency of the position it will
// be recorded on, and returns a copy with every currency and the rate resolved,
// so the caller writes the settled values rather than re-deriving them.
//
// Three things are refused, all because they would corrupt a cost basis rather
// than fail:
//
//   - a rate other than 1 between a currency and itself, which is a typo no
//     arithmetic can absorb: it scales the position's cost by whatever number
//     was typed and nothing downstream can tell.
//   - a missing rate between two different currencies. The tempting fallback is
//     to look up today's rate, and it is exactly wrong: today's rate applied to
//     a trade from December is the re-translation this whole column exists to
//     stop. A caller who does not know the historical rate has to say so, and
//     the client fills the field from the stored rate as a suggestion the owner
//     can correct against their confirmation.
//   - a fee currency that is neither of the two the transaction has. A fee in
//     some third currency has no rate on this row to reach the position with,
//     and storing it would mean carrying a second historical rate for an amount
//     that is a rounding error next to the trade.
//
// FeesCurrency defaults to the trade currency, not to the account's. That is
// what every version of this app did while the two were forced to be equal, and
// what a spreadsheet's single currency column means for the row it tags, so the
// default changes no existing behaviour. A broker that bills the account
// instead has to say so — which the forms do, explicitly, on the only screens
// where the two can differ.
func (in TransactionInput) Validate(costCurrency money.Currency) (TransactionInput, error) {
	out := in

	if out.Currency == money.XXX {
		out.Currency = costCurrency
	}

	out.FXRate = in.Rate()

	if !out.FXRate.IsPos() {
		return in, fmt.Errorf("%w: rate must be greater than zero, got %s", ErrTransactionFXRate, out.FXRate.String())
	}

	switch {
	case out.Currency == costCurrency:
		if !out.FXRate.Equal(decimal.One) {
			return in, fmt.Errorf("%w: %s does not convert into itself at %s", ErrTransactionFXRate, out.Currency, out.FXRate.String())
		}
	case in.FXRate.IsZero():
		return in, fmt.Errorf("%w: the trade is in %s and the position costs in %s, so fxRate is required", ErrTransactionFXRate, out.Currency, costCurrency)
	}

	if out.FeesCurrency == money.XXX {
		out.FeesCurrency = out.Currency
	}

	if out.FeesCurrency != out.Currency && out.FeesCurrency != costCurrency {
		return in, fmt.Errorf("%w: fees in %s belong to neither the trade (%s) nor the position (%s)", ErrTransactionFeesCurrency, out.FeesCurrency, out.Currency, costCurrency)
	}

	return out, nil
}

// FeesInCostCurrency is the Go side of the transaction_fees_in_cost SQL
// function: the commission expressed in the position's currency, which is the
// only form in which it can be added to anything else on the position.
//
// The two implementations exist because the growth series is computed in SQL
// and the activity alert in Go; they are kept adjacent in name so a change to
// one is visibly a change the other needs.
func (in TransactionInput) FeesInCostCurrency() money.Money {
	if in.FeesCurrency == in.Currency {
		return in.Fees.MulDecimal(in.Rate())
	}

	return in.Fees
}

// ImportTransactionRow is one validated spreadsheet row, ready to be
// persisted as an asset + portfolio entry + transaction.
type ImportTransactionRow struct {
	RowNumber int
	Ticker    string
	AssetName string
	AssetType market.AssetType
	Type      TransactionType
	Quantity  decimal.Decimal
	Price     money.Money
	Fees      money.Money
	Currency  money.Currency
	// FXRate and CostCurrency carry the same meaning they do on a transaction:
	// the row settled into CostCurrency at FXRate. CostCurrency comes from the
	// upload's defaults rather than the row, because a statement is one account.
	FXRate       decimal.Decimal
	CostCurrency money.Currency
	Date         time.Time
	Notes        string
}

// Input turns an import row into the same shape the HTTP writers use, so the
// importer is checked by the rule the API is checked by instead of a second
// copy of it. Fees follow the trade currency, which is what the file's own
// currency column means for the amounts on the row it tags.
func (r ImportTransactionRow) Input() TransactionInput {
	return TransactionInput{
		Type:            r.Type,
		Quantity:        r.Quantity,
		Price:           r.Price,
		Currency:        r.Currency,
		FXRate:          r.FXRate,
		Fees:            r.Fees,
		FeesCurrency:    r.Currency,
		TransactionDate: r.Date,
		Notes:           r.Notes,
	}
}

type Snapshot struct {
	ID           uuid.UUID      `json:"id"`
	PortfolioID  uuid.UUID      `json:"portfolioId"`
	SnapshotDate time.Time      `json:"snapshotDate"`
	TotalValue   money.Money    `json:"totalValue"`
	Currency     money.Currency `json:"currency"`
	// Allocation is a JSON object, so it is held raw: as []byte it would reach
	// a client base64-encoded.
	Allocation       json.RawMessage `json:"allocation"`
	TotalGainLoss    money.Money     `json:"totalGainLoss"`
	TotalGainLossPCT money.Money     `json:"totalGainLossPCT"`
	CreatedAt        time.Time       `json:"createdAt"`
	Portfolio        Portfolio       `json:"portfolio,omitzero"`
}

// AllocationItem is the result of grouping portfolio_entries by category.
type AllocationItem struct {
	Category             market.AssetType `json:"category"`
	MarketValue          string           `json:"marketValue"`
	Currency             money.Currency   `json:"currency"`
	PositionsUnconverted int64            `json:"positionsUnconverted"`
}

// AssetHolding is one asset totalled across every portfolio the user owns.
//
// It answers a question no other view does: "how much of X do I have?", without
// asking which portfolio it sits in. The per-portfolio holdings only ever add
// up within their own portfolio, and the allocation folds everything down to
// eight asset types, so an asset split across three portfolios had no single
// row anywhere in the app.
//
// Quantity is a sum of units and therefore only means something per asset —
// never across rows. MarketValue is what does compare, and it is in
// DisplayCurrency for every row, which is what makes a share of the total
// meaningful (same rule as AllocationItem).
type AssetHolding struct {
	AssetID   uuid.UUID        `json:"assetId"`
	Ticker    string           `json:"ticker"`
	Name      string           `json:"name"`
	AssetType market.AssetType `json:"assetType"`
	Exchange  string           `json:"exchange"`
	// Currency the asset is quoted in, which is what MarketPrice is in. It is
	// not DisplayCurrency: the price stays in its own currency, the value is
	// converted.
	Currency money.Currency `json:"currency"`
	// Quantity is the units held across every portfolio, as text for the same
	// reason the amounts are: the decimal engine, not float64, is what reads it.
	Quantity string `json:"quantity"`
	// MarketPrice is the per-unit price behind MarketValue, in Currency. Empty
	// when PriceSource is cost — there was no price, and the entries' own costs
	// differ from one another, so no single number stands for the asset.
	MarketPrice string `json:"marketPrice"`
	// MarketValue is the whole position, in DisplayCurrency.
	MarketValue     string         `json:"marketValue"`
	DisplayCurrency money.Currency `json:"displayCurrency"`
	// Portfolios is how many of the user's portfolios hold this asset — the
	// count the consolidated row hides by construction.
	Portfolios  int64       `json:"portfolios"`
	PriceSource PriceSource `json:"priceSource"`
	// PriceProvider names the market-data provider whose key produced
	// MarketPrice, and PriceFetchedAt is when it did. Both are empty unless
	// PriceSource is own: a manual catalog price and a cost basis come from no
	// provider.
	//
	// PriceSource alone says the price is the user's own, which is enough to
	// know it is not a cost basis and not enough to act on. Deciding whether to
	// re-ask — and whom — needs the name and the timestamp: "Finnhub, three days
	// ago" is a stale quote worth refreshing, and the same number fetched a
	// minute ago is simply what the market says.
	PriceProvider  string     `json:"priceProvider,omitempty"`
	PriceFetchedAt *time.Time `json:"priceFetchedAt,omitempty"`
	// PositionsUnconverted counts this asset's entries that had no rate to
	// DisplayCurrency and are therefore added at face value; same meaning as in
	// AllocationItem and the summary.
	PositionsUnconverted int64 `json:"positionsUnconverted"`
}

// PlatformStats is the result of joining investment_sources with portfolio_entries stats.
type PlatformStats struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SourceType  SourceType `json:"sourceType"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	// Investments counts the platform's open positions — entries still holding
	// something. A position sold in full is not one of them, the same rule the
	// holdings and the allocation apply.
	Investments int64 `json:"investments"`
	// Assets and Portfolios are what those positions are spread over: the same
	// ten positions are a different account when they are ten companies than
	// when they are one company held in ten portfolios.
	Assets     int64 `json:"assets"`
	Portfolios int64 `json:"portfolios"`
	// TotalValue is the cost basis: what the owner put in, at weighted-average
	// cost. The name predates there being anything else to compare it against
	// and is kept because clients read it; MarketValue is what the same
	// positions are worth now, in the same currency and over the same rows, so
	// the two subtract into a real gain.
	//
	// Both are the cost and value of what is *still held*: quantity is the net
	// left after sales, so a position sold down to half carries half its cost.
	// Neither includes commissions — the average cost the trigger writes is a
	// price, not a cash outlay — and neither counts what a sale realised.
	TotalValue           string         `json:"totalValue"`
	MarketValue          string         `json:"marketValue"`
	DisplayCurrency      money.Currency `json:"displayCurrency"`
	PositionsUnconverted int64          `json:"positionsUnconverted"`
	// Where MarketValue came from, partitioning Investments exactly. A position
	// with no price is carried at its own cost and contributes zero to the
	// gain, so a platform reporting no gain at all is either flat or unpriced,
	// and only these three counts separate the two.
	PositionsPricedOwn    int64 `json:"positionsPricedOwn"`
	PositionsPricedManual int64 `json:"positionsPricedManual"`
	PositionsAtCost       int64 `json:"positionsAtCost"`
}

// SummaryView is the result of joining portfolios + risks + portfolio_summary view.
type SummaryView struct {
	ID                    uuid.UUID      `json:"id"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	Type                  Type           `json:"type"`
	BaseCurrency          money.Currency `json:"baseCurrency"`
	IsDefault             bool           `json:"isDefault"`
	RiskID                uuid.UUID      `json:"riskId"`
	RiskName              string         `json:"riskName"`
	TotalPositions        int64          `json:"totalPositions"`
	TotalCostBase         string         `json:"totalCostBase"`
	TotalMarketValue      string         `json:"totalMarketValue"`
	TotalGainLoss         string         `json:"totalGainLoss"`
	TotalGainLossPct      string         `json:"totalGainLossPct"`
	CreatedAt             time.Time      `json:"createdAt"`
	DisplayCurrency       money.Currency `json:"displayCurrency"`
	FXConverted           bool           `json:"fxConverted"`
	PositionsPricedOwn    int64          `json:"positionsPricedOwn"`
	PositionsPricedManual int64          `json:"positionsPricedManual"`
	PositionsAtCost       int64          `json:"positionsAtCost"`
	PositionsUnconverted  int64          `json:"positionsUnconverted"`
}

type SnapshotRow struct {
	PortfolioID      uuid.UUID
	BaseCurrency     money.Currency
	TotalMarketValue string
	TotalCostBase    string
	TotalGainLoss    string
	TotalGainLossPct string
	// Allocation is the day's composition, as a JSON object mapping each
	// market.AssetType to what the portfolio held of it, in BaseCurrency and as
	// text — the same shape and the same reading the totals above get. It comes
	// out of the same statement as TotalMarketValue so the parts add up to it
	// by construction; see GetAllPortfolioSummaryRows.
	Allocation string
}

type GrowthPoint struct {
	Date          time.Time
	TotalValue    string
	TotalCostBase string
	GainLoss      string
	GainLossPct   string
	// Currency the three amounts above are expressed in. The aggregate series
	// converts every portfolio into it; a single portfolio's series is already
	// in its own base currency.
	Currency money.Currency
	// How many portfolios went into this date without a rate to convert them,
	// and are therefore added at face value. Zero is the normal case.
	PortfoliosUnconverted int64
	// NetFlow is the money the owner put in (positive) or took out (negative)
	// between the previous point of the series and this one, in Currency. It is
	// what has to be netted out of the change in TotalValue to leave a return:
	// a deposit raises the value without anyone having earned anything. The
	// sign convention lives in the transaction_cash_flow SQL function, and the
	// first point of a series carries whatever fell on or before its own date,
	// which no subperiod uses.
	NetFlow string
}

// GrowthSummary carries two different readings of the same series and they must
// not be confused. InitialValue/CurrentValue/TotalGrowthPct describe how the
// value moved between the first snapshot and the last, which counts money put
// in as growth. GainLoss/GainLossPct are the profit of the latest point —
// market value minus invested capital — which is the actual return.
type GrowthSummary struct {
	FirstDate      time.Time
	InitialValue   string
	CurrentValue   string
	TotalGrowthPct string
	GainLoss       string
	GainLossPct    string
	Currency       money.Currency
}

// PortfolioValuePoint is what one portfolio was worth on one snapshot date,
// in its own base currency. Date is carried alongside the amount because the
// snapshot found is the most recent one on or before the date asked for, not
// necessarily that date itself — a caller reporting "since last week" has to
// be able to say which day it actually compared against.
type PortfolioValuePoint struct {
	PortfolioID uuid.UUID
	Date        time.Time
	TotalValue  string
}
