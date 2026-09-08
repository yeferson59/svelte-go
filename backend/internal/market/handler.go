package market

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

type handler struct {
	service *service
	// holdings answers "which assets does this user own", supplied by the
	// composition root from the portfolio module. market must not import
	// portfolio, so it is consumed through the interface declared here.
	holdings Holdings
}

func (h *handler) ImportAssets(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.service.ImportAssetsFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return httpx.FromDomain(c, err, "Error importing assets", "Could not import the uploaded assets")
	}

	return httpx.OK(c, "Assets imported", "Spreadsheet imported successfully", result)
}

// CreateAsset adds an asset to the catalog, curating it when an admin asks and
// contributing it when anybody else does.
//
// One route, two behaviours, instead of two routes: from the client's side this
// is the same request — "put this instrument in the catalog" — and the caller
// does not get to choose which of the two it becomes.
func (h *handler) CreateAsset(c fiber.Ctx) error {
	userID, _, role, err := httpx.Identity(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid identity", err.Error())
	}

	req, err := httpx.Bind[CreateAssetRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	assetType := AssetType(req.AssetType)

	var asset Asset
	if role == httpx.RoleAdmin {
		asset, err = h.service.CreateAsset(c, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	} else {
		asset, err = h.service.ContributeAsset(c, userID, req.Ticker, req.Name, assetType, req.Exchange, req.Currency)
	}

	if err != nil {
		return httpx.FromDomain(c, err, "Error creating asset", assetFailureDetail(err, "No se pudo crear el activo"))
	}

	return httpx.Success(c, fiber.StatusCreated, "Asset created", "Asset created successfully", asset)
}

// UpdateAsset rewrites a catalog row: its ticker, name, type, exchange,
// currency, who sees it, and the manual price.
//
// Admin-only, unlike CreateAsset above, and the guard is on the route. The two
// are not the same request wearing different roles: creating names an
// instrument that was missing, while this reaches an existing row by id — one
// that other users may hold positions in — and can rename it, re-denominate it
// or take it off the shared catalog.
func (h *handler) UpdateAsset(c fiber.Ctx) error {
	assetID, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid asset ID", err.Error())
	}

	req, err := httpx.Bind[UpdateAssetRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	asset, err := h.service.UpdateAsset(c, assetID, AssetUpdate{
		Ticker:    req.Ticker,
		Name:      req.Name,
		AssetType: AssetType(req.AssetType),
		Exchange:  req.Exchange,
		Currency:  req.Currency,
		IsCurated: req.IsCurated,
		Price:     req.Price,
	})
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating asset", assetFailureDetail(err, "No se pudo actualizar el activo"))
	}

	return httpx.OK(c, "Asset updated", "Asset updated successfully", asset)
}

func (h *handler) GetExchangeRates(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return httpx.InternalServerError(c, "", "paginate info not found")
	}

	rates, err := h.service.GetExchangeRates(c, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", rates)
}

// GetLatestExchangeRates serves the shared rates to any signed-in user, where
// the listing above stays admin-only.
//
// The two answer different questions and only one of them is administration:
// this is the dashboard asking what a dollar is worth so it can label the
// figures it already converts, and the rates it returns are public data the
// application fetched without anybody's key. Gating it behind the admin role
// would hide from users the provenance of numbers they are being shown.
func (h *handler) GetLatestExchangeRates(c fiber.Ctx) error {
	rates, err := h.service.GetLatestExchangeRates(c)
	if err != nil {
		return httpx.FromDomain(c, err, "", "")
	}

	return httpx.OK(c, "", "", rates)
}

// RefreshExchangeRates pulls the public feed on demand.
//
// The scheduled job is what keeps the rates current; this exists for the two
// moments the cadence cannot cover — a fresh deployment, whose first scheduled
// run is a full interval away, and an operator who needs today's rate now.
func (h *handler) RefreshExchangeRates(c fiber.Ctx) error {
	rates, err := h.service.RefreshPublicRates(c)
	if err != nil {
		return httpx.FromDomain(c, err, "Error refreshing exchange rates", "Could not refresh the public exchange rates")
	}

	return httpx.OK(c, "Exchange rates refreshed", "Public exchange rates refreshed successfully", rates)
}

func (h *handler) CreateExchangeRate(c fiber.Ctx) error {
	req, err := httpx.Bind[CreateExchangeRateRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.CreateExchangeRate(c, req.FromCurrency, req.ToCurrency, req.Rate)
	if err != nil {
		return httpx.FromDomain(c, err, "Error creating exchange rate", "Could not create exchange rate")
	}

	return httpx.Success(c, fiber.StatusCreated, "Exchange rate created", "Exchange rate created successfully", rate)
}

func (h *handler) UpdateExchangeRate(c fiber.Ctx) error {
	id, err := httpx.ParamUUID(c, "id")
	if err != nil {
		return httpx.BadRequest(c, "Invalid exchange rate ID", err.Error())
	}

	req, err := httpx.Bind[UpdateExchangeRateRequestDTO](c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid request", err.Error())
	}

	rate, err := h.service.UpdateExchangeRate(c, id, req.Rate)
	if err != nil {
		return httpx.FromDomain(c, err, "Error updating exchange rate", "Could not update exchange rate")
	}

	return httpx.OK(c, "Exchange rate updated", "Exchange rate updated successfully", rate)
}

func (h *handler) ImportExchangeRates(c fiber.Ctx) error {
	data, filename, err := readImportFile(c)
	if err != nil {
		return httpx.BadRequest(c, "Invalid file", err.Error())
	}

	result, err := h.service.ImportExchangeRatesFromFile(c, data, filename, c.FormValue("sheet"))
	if err != nil {
		return httpx.FromDomain(c, err, "Error importing exchange rates", "Could not import the uploaded exchange rates")
	}

	return httpx.OK(c, "Exchange rates imported", "Spreadsheet imported successfully", result)
}

// maxImportFileSize bounds uploaded spreadsheets; classic personal trackers
// with a few thousand rows stay well under this.
const maxImportFileSize = 8 << 20 // 8 MiB

func readImportFile(c fiber.Ctx) ([]byte, string, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, "", errors.New("missing file: attach the spreadsheet in the \"file\" field")
	}
	if fileHeader.Size > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer func(f multipart.File) { _ = f.Close() }(file)

	data, err := io.ReadAll(io.LimitReader(file, maxImportFileSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxImportFileSize {
		return nil, "", errors.New("file too large: the maximum size is 8 MB")
	}
	return data, fileHeader.Filename, nil
}
