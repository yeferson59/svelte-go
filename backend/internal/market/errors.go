package market

import (
	"errors"
	"fmt"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// Domain sentinel errors for the "not found" family, tagged with their HTTP
// Kind so httpx.FromDomain maps them by type rather than by the text of the
// message — the 404 survives a caller wrapping the error with a message that
// happens to contain "failed"/"invalid".
//
// Messages are preserved verbatim from the persistence layer so response
// bodies and any errors.Is/message assertions keep working.
var (
	ErrAssetNotFound        = httpx.AsNotFound(errors.New("asset not found"))
	ErrExchangeRateNotFound = httpx.AsNotFound(errors.New("exchange rate not found"))
	ErrCredentialNotFound   = httpx.AsNotFound(errors.New("market data credential not found"))
)

// Catalog input errors. Both creation paths — the operator's and the user's —
// validate through normalizeAssetInput and return these, so the same bad ticker
// is answered the same way whoever sent it.
var (
	errAssetTickerRequired  = httpx.AsBadRequest(errors.New("el ticker es obligatorio"))
	errAssetTickerTooLong   = httpx.AsBadRequest(errors.New("el ticker supera el máximo de caracteres"))
	errAssetExchangeTooLong = httpx.AsBadRequest(errors.New("el mercado supera el máximo de caracteres"))
	errAssetTypeInvalid     = httpx.AsBadRequest(errors.New("el tipo de activo debe ser uno de: stock, etf, crypto, bond, cash, real_estate, commodity, other"))
	errAssetCurrencyInvalid = httpx.AsBadRequest(errors.New("la moneda debe ser un código ISO de 3 letras"))

	// ErrAssetQuotaExceeded is a 429 rather than a 403: the user is allowed to
	// contribute assets, just not this many more today.
	ErrAssetQuotaExceeded = httpx.AsTooManyRequests(fmt.Errorf("has añadido demasiados activos nuevos en las últimas 24 horas (máximo %d)", maxContributedAssetsPerDay))
)

// Exchange rate input errors. The admin endpoints validate through the same
// ISO 4217 table and positivity rule the spreadsheet importer uses, so a rate
// that no conversion could apply is rejected where it is entered.
var (
	errExchangeRateCurrencyInvalid = httpx.AsBadRequest(errors.New("la moneda debe ser un código ISO 4217 válido"))
	errExchangeRateInvalid         = httpx.AsBadRequest(errors.New("la tasa de cambio debe ser mayor que 0"))
)

// assetFailureDetail returns the message only when this package authored it.
// Anything else — a driver error, a constraint name — is replaced, so a failed
// insert cannot echo the schema back over the wire.
func assetFailureDetail(err error) string {
	for _, domain := range []error{
		errAssetTickerRequired, errAssetTickerTooLong, errAssetExchangeTooLong,
		errAssetTypeInvalid, errAssetCurrencyInvalid, ErrAssetQuotaExceeded,
	} {
		if errors.Is(err, domain) {
			return err.Error()
		}
	}

	return "No se pudo crear el activo"
}

// BYO-key errors. A user with no key is not a failure of the application, so
// ErrNoCredentials is a 400 the UI turns into "add your key in settings"
// rather than a 500.
var (
	// Wraps the marketdata sentinel so callers can match either: this is the
	// same condition, carrying an HTTP kind. Without the wrap the two would be
	// distinct errors for one state.
	ErrNoCredentials   = httpx.AsBadRequest(fmt.Errorf("no market data credential configured: %w", marketdata.ErrNoCredentials))
	ErrInvalidProvider = httpx.AsBadRequest(errors.New("unknown market data provider"))
	ErrInvalidAPIKey   = httpx.AsBadRequest(errors.New("the provider rejected this API key"))
	// ErrProviderUnavailable means the provider could not be reached or answered
	// something we cannot classify — a timeout, a 5xx, a malformed body.
	//
	// It exists to keep that case apart from ErrInvalidAPIKey. Collapsing the two
	// told the user their key was rejected during an outage, and, worse, let the
	// verify endpoint persist status='invalid', which the sync queries then skip
	// for good: one bad afternoon at the provider would silently retire a working
	// key. Untagged, so it maps to 500 like every other upstream failure.
	ErrProviderUnavailable = errors.New("the market data provider could not be reached")
	ErrKeyEncryptionOff    = errors.New("market: credential encryption is not configured")
)

// Single-asset refresh errors. The daily sync needs none of these: it runs over
// every holding and treats each failure as one row it could not fill, so a
// savings account nobody quotes and a symbol a provider does not carry are both
// things it skips without a word. Somebody who pressed «Actualizar» on one
// asset asked a question, and each of these is a different answer to it.
var (
	// ErrAssetNotQuotable is about the asset: cash, real estate and the other
	// types no market-data provider prices.
	ErrAssetNotQuotable = httpx.AsBadRequest(errors.New("this asset is not quoted by any market data provider"))
	// ErrAssetNotCovered is about the pairing: the provider works, the asset is
	// quotable, and this provider has no data for it. Another key may.
	ErrAssetNotCovered = httpx.AsBadRequest(errors.New("this provider has no data for this asset"))
	// ErrProviderQuotaSpent is a 429 because that is what it is: the key is
	// fine and the answer is to come back later.
	ErrProviderQuotaSpent = httpx.AsTooManyRequests(errors.New("the quota of your key with this provider is spent"))
)

// refreshFailureDetail is what the user reads when a single-asset refresh
// fails.
//
// It answers in Spanish, unlike credentialFailureDetail, because these strings
// are shown verbatim beside the asset that was pressed. Settings can rewrite the
// English sentinels on the client because a save has two outcomes; a refresh has
// six, and mapping them from a status code alone is not possible.
func refreshFailureDetail(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrAssetNotQuotable):
		return "Ningún proveedor de precios cotiza este tipo de activo."
	case errors.Is(err, ErrAssetNotCovered):
		return "Este proveedor no tiene datos de este activo. Prueba con otra de tus claves."
	case errors.Is(err, ErrProviderQuotaSpent):
		return "La cuota de tu clave con este proveedor se agotó. Vuelve a intentarlo más tarde."
	case errors.Is(err, ErrInvalidAPIKey):
		return "El proveedor rechazó tu clave. Revísala en Ajustes."
	case errors.Is(err, ErrCredentialNotFound), errors.Is(err, ErrNoCredentials), errors.Is(err, ErrKeyEncryptionOff):
		return "No tienes una clave guardada para este proveedor."
	case errors.Is(err, ErrInvalidProvider):
		return "Ese proveedor de precios no existe."
	case errors.Is(err, ErrAssetNotFound):
		return "Este activo ya no está en el catálogo."
	default:
		return "El proveedor no respondió. Inténtalo de nuevo en unos minutos."
	}
}

// isDomainCredentialError reports whether err is one this package authored, and
// is therefore safe to show the user. Anything else — a provider's own text, a
// transport failure — is replaced by a generic message before it reaches a
// response body.
func isDomainCredentialError(err error) bool {
	for _, domain := range []error{ErrNoCredentials, ErrInvalidProvider, ErrInvalidAPIKey, ErrProviderUnavailable, ErrCredentialNotFound} {
		if errors.Is(err, domain) {
			return true
		}
	}

	return false
}
