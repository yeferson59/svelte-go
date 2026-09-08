// Package market owns the asset catalog and exchange-rate domains: the request
// DTOs, entities, persistence, services and HTTP handlers for both.
package market

import (
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type CreateAssetRequestDTO struct {
	Ticker    string         `json:"ticker"    validate:"required"`
	Name      string         `json:"name"      validate:"required"`
	AssetType string         `json:"assetType" validate:"required"`
	Exchange  string         `json:"exchange"`
	Currency  money.Currency `json:"currency"  validate:"required"`
}

// UpdateAssetRequestDTO is a catalog row as the operator wants it to read from
// now on: the identifying fields travel whole, so an edit that only changes the
// name still sends the ticker it is keeping.
//
// The last two are pointers because they are the fields whose zero value is a
// legitimate instruction. A body with no isCurated leaves the audience alone,
// and one with `false` hides the row from everybody but its contributors; a
// body with no price leaves the manual price alone, which is not the same as
// clearing it.
type UpdateAssetRequestDTO struct {
	Ticker    string         `json:"ticker"    validate:"required"`
	Name      string         `json:"name"      validate:"required"`
	AssetType string         `json:"assetType" validate:"required"`
	Exchange  string         `json:"exchange"`
	Currency  money.Currency `json:"currency"  validate:"required"`
	IsCurated *bool          `json:"isCurated"`
	Price     *money.Money   `json:"price"`
}

type CreateExchangeRateRequestDTO struct {
	FromCurrency money.Currency  `json:"fromCurrency" validate:"required"`
	ToCurrency   money.Currency  `json:"toCurrency"   validate:"required"`
	Rate         decimal.Decimal `json:"rate"         validate:"required"`
}

type UpdateExchangeRateRequestDTO struct {
	Rate decimal.Decimal `json:"rate" validate:"required"`
}

// SaveCredentialRequestDTO carries a user's own provider API key.
//
// It has no response counterpart on purpose: a stored key is never sent back,
// so nothing that leaves this module can carry one.
type SaveCredentialRequestDTO struct {
	APIKey string `json:"apiKey" validate:"required,min=8,max=256"`
}

// RefreshAssetPriceRequestDTO names the key one asset must be re-quoted with.
//
// The provider is required and has no default on purpose: the whole point of
// the request is that the caller chose one, and a body that left it out would
// quietly become a fallback run whose result is attributed to whichever
// provider happened to answer.
type RefreshAssetPriceRequestDTO struct {
	Provider string `json:"provider" validate:"required"`
}

// SyncResultDTO is what POST /market/sync returns: the prices and the rates the
// caller's own keys fetched.
//
// Both halves are named because a sync is not just prices. A holding quoted in
// a currency other than its portfolio's needs a rate to be worth anything, and
// under BYO-key nobody else's rate may be used, so a run that refreshed prices
// but no rates has done half the job — and the response should say which half.
type SyncResultDTO struct {
	Prices []UserAssetPrice   `json:"prices"`
	Rates  []UserExchangeRate `json:"rates"`
}
