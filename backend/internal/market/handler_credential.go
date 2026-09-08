package market

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The responses below carry Credential values, which have no field for the API
// key or the ciphertext. That is the guarantee: there is no shape of this
// handler that can serve a key back, not even to its owner.

func (h *handler) ListCredentials(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	creds, err := h.service.ListCredentials(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Error reading credentials", "Could not read your market data keys")
	}

	return httpx.OK(c, "", "", creds)
}

func (h *handler) SaveCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	provider := ProviderID(c.Params("provider"))

	req, err := httpx.Bind[SaveCredentialRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	cred, err := h.service.SaveCredential(c, userID, provider, req.APIKey)
	if err != nil {
		// err may name the provider but never the key: every provider error is
		// built through marketdata.Errorf, which scrubs it.
		return httpx.FromDomain(c, err, "Could not save the key", credentialFailureDetail(err))
	}

	return httpx.OK(c, "Key saved", "The key was verified against the provider and stored encrypted", cred)
}

func (h *handler) VerifyCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	cred, err := h.service.VerifyCredential(c, userID, ProviderID(c.Params("provider")))
	if err != nil {
		return httpx.FromDomain(c, err, "Could not verify the key", credentialFailureDetail(err))
	}

	return httpx.OK(c, "Key verified", "", cred)
}

func (h *handler) DeleteCredential(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	if err := h.service.DeleteCredential(c, userID, ProviderID(c.Params("provider"))); err != nil {
		return httpx.FromDomain(c, err, "Could not delete the key", "The key could not be deleted")
	}

	return httpx.OK(c, "Key deleted", "", nil)
}

// onDemandSyncBudget bounds how long a user is made to wait on the sync button.
//
// The sync paces its calls to fit personal free-tier quotas — 13s apart with an
// Alpha Vantage key — so a portfolio of any size runs for minutes. Rather than
// hold the request open for all of it, the work is cut off here and whatever was
// fetched is returned; the daily job picks up the rest. The holdings come back
// most-recently-traded first, so the budget is spent on what the user is most
// likely looking at.
const onDemandSyncBudget = 60 * time.Second

// SyncMarketData refreshes the caller's own holdings with the caller's own
// keys. It replaces the admin-only global sync, which no longer has a key to
// run under.
func (h *handler) SyncMarketData(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	assetIDs, err := h.holdings.HeldAssetIDs(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Could not read your holdings", "")
	}

	pairs, err := h.holdings.RequiredCurrencyPairs(c, userID)
	if err != nil {
		return httpx.FromDomain(c, err, "Could not read your holdings", "")
	}

	ctx, cancel := context.WithTimeout(c.Context(), onDemandSyncBudget)
	defer cancel()

	prices, errs := h.service.SyncAssetsForUser(ctx, userID, assetIDs)
	if len(errs) > 0 && len(prices) == 0 {
		return httpx.FromDomain(c, errs[0], "Market data sync failed", credentialFailureDetail(errs[0]))
	}

	// The rates matter as much as the prices: without them a holding quoted in
	// another currency has nothing to convert through, and the shared table no
	// longer carries a rate anybody can fall back on. Their failures do not fail
	// the request — the prices above already succeeded.
	rates, _ := h.service.SyncRatesForUser(ctx, userID, pairs)

	return httpx.OK(c, "Market data synced", "", SyncResultDTO{Prices: prices, Rates: rates})
}

// RefreshAssetPrice re-quotes one asset with one key of the caller's own.
//
// SyncMarketData above refreshes every holding and lets the chain pick who
// answers; this is the same act narrowed to the asset somebody is looking at,
// with the provider named. It carries the credential limiter for the same
// reason sync does: each call spends the user's own quota.
func (h *handler) RefreshAssetPrice(c fiber.Ctx) error {
	userID, _, _, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	assetID, err := httpx.ParamUUID(c, "assetId")
	if err != nil {
		return httpx.BadRequest(c, "Invalid asset ID", err.Error())
	}

	req, err := httpx.Bind[RefreshAssetPriceRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	price, err := h.service.RefreshAssetPrice(c, userID, assetID, ProviderID(req.Provider))
	if err != nil {
		return httpx.FromDomain(c, err, "Could not refresh the price", refreshFailureDetail(err))
	}

	return httpx.OK(c, "Price updated", "", price)
}

// credentialFailureDetail keeps provider text out of the response body unless
// it is one of our own domain errors.
//
// The old handlers returned errs[0].Error() straight to the client. With the
// operator's key that was merely untidy; with a user's key it would be a way to
// echo provider output — and the key rides in Alpha Vantage's URL — back over
// the wire. Only the messages this package authored are returned now.
func credentialFailureDetail(err error) string {
	switch {
	case err == nil:
		return ""
	case isDomainCredentialError(err):
		return err.Error()
	default:
		return "The market data provider could not be reached. Try again in a few minutes."
	}
}
