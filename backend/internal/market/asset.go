package market

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

// Column limits mirror the assets table (ticker VARCHAR(20), name VARCHAR(255),
// exchange VARCHAR(100)).
const (
	maxTickerLen    = 20
	maxAssetNameLen = 255
	maxExchangeLen  = 100
)

// AssetType is the kind of a tradable asset.
type AssetType string

const (
	Stock      AssetType = "stock"
	ETF        AssetType = "etf"
	Crypto     AssetType = "crypto"
	Bond       AssetType = "bond"
	Cash       AssetType = "cash"
	RealEstate AssetType = "real_estate"
	Commodity  AssetType = "commodity"
	Other      AssetType = "other"
)

func (a AssetType) IsValid() bool {
	switch a {
	case Stock, ETF, Crypto, Bond, Cash, RealEstate, Commodity, Other:
		return true
	default:
		return false
	}
}

// Asset is a tradable instrument in the catalog. Owned by the market module;
// the portfolio module references it (entries hold an Asset) but does not own
// its lifecycle.
type Asset struct {
	ID             uuid.UUID      `json:"id"`
	Ticker         string         `json:"ticker"`
	Name           string         `json:"name"`
	AssetType      AssetType      `json:"assetType"`
	Exchange       string         `json:"exchange"`
	Currency       money.Currency `json:"currency"`
	CurrentPrice   *money.Money   `json:"currentPrice"`
	PriceUpdatedAt *time.Time     `json:"priceUpdatedAt"`
	// IsCurated marks a row the operator vouches for, which is what makes it
	// visible to every user. A contributed row is only visible to the users who
	// contributed it, so the flag doubles as the "everyone" audience.
	IsCurated bool      `json:"isCurated"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AssetUpdate is the new state of a catalog row, as an operator edit leaves it.
//
// A struct rather than eight parameters because two of the fields are optional
// in a way an argument list hides: a nil IsCurated keeps the audience the row
// already has, and a nil Price keeps the manual price. Everything else is
// replaced with what is here, so an edit is read as "this is the row now" and
// not as a patch whose omissions have to be guessed at.
type AssetUpdate struct {
	Ticker    string
	Name      string
	AssetType AssetType
	Exchange  string
	Currency  money.Currency
	// IsCurated changes who sees the row: curating it publishes it to every
	// user, un-curating it puts it back to the users who contributed it. Nil
	// leaves it as it is.
	IsCurated *bool
	// Price is the manual fallback price, written with PriceUpdatedAt set to
	// now. Nil leaves both alone — except when the currency changes, which
	// invalidates whatever number is stored (see UpdateAsset).
	Price *money.Money
}

// CatalogView is the audience a catalog read is served for: the curated rows,
// plus the ones ViewerID contributed. All lifts the scope to the whole table
// and is set only for admins, who moderate what users contribute and would
// otherwise be unable to see it.
//
// It is a type rather than a bare uuid parameter because the two fields must
// travel together — a viewer id with no All flag reads as "this user only",
// which is exactly the wrong answer for an admin.
type CatalogView struct {
	ViewerID uuid.UUID
	All      bool
}

// normalizeCurrencyCode validates a currency spelling against gofinance's ISO
// 4217 table and returns the Currency it names. Length alone was never the test
// that mattered: every amount this module stores is tagged with a
// money.Currency, and GetCurrencyFromISOCode is what has to accept the code for
// that to work. A three-letter string it rejects — "DOL", "ABC" — would reach
// the database as the zero currency, which prints as XXX and costs the asset
// its price.
//
// The importers are the only callers: everywhere else the currency arrives as a
// Currency already, decoded by the money package's own UnmarshalText.
func normalizeCurrencyCode(raw string) (money.Currency, bool) {
	cur, err := money.GetCurrencyFromISOCode(raw)
	if err != nil {
		return money.XXX, false
	}

	return cur, true
}

var categorySynonyms = map[string]AssetType{
	"stock": Stock, "stocks": Stock, "accion": Stock, "acciones": Stock,
	"equity": Stock, "equities": Stock,
	"etf": ETF, "etfs": ETF, "fondo": ETF, "fondos": ETF,
	"fund": ETF, "funds": ETF, "fondo indexado": ETF, "index fund": ETF,
	"crypto": Crypto, "cripto": Crypto, "criptomoneda": Crypto,
	"criptomonedas": Crypto, "cryptocurrency": Crypto, "criptos": Crypto,
	"bond": Bond, "bonds": Bond, "bono": Bond, "bonos": Bond,
	"renta fija": Bond, "cdt": Bond, "fixed income": Bond,
	"cash": Cash, "efectivo": Cash, "liquidez": Cash, "dinero": Cash,
	"real estate": RealEstate, "real_estate": RealEstate, "inmueble": RealEstate,
	"inmuebles": RealEstate, "bienes raices": RealEstate, "reit": RealEstate, "fibra": RealEstate,
	"commodity": Commodity, "commodities": Commodity, "materia prima": Commodity,
	"materias primas": Commodity, "oro": Commodity, "gold": Commodity, "plata": Commodity,
	"other": Other, "otro": Other, "otros": Other,
}

// NormalizeAssetType maps a free-form category label (accent/case-insensitive)
// to a known AssetType. It is shared with the portfolio importer, which needs
// the same mapping when a transaction file carries a category column.
func NormalizeAssetType(raw string) (AssetType, bool) {
	if c, ok := categorySynonyms[spreadsheet.NormKey(raw)]; ok {
		return c, true
	}
	return "", false
}
