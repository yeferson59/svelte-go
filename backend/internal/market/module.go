package market

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// This module is built in two steps, and the order matters at the composition
// root — the same shape user and marketing already use.
//
// NewService depends on no other domain module, so it runs first and its result
// is handed to portfolio, which reads the asset catalog through it. New then
// completes the module with the route guards and with portfolio's holdings,
// which only exist once portfolio is built. Splitting the two is what lets the
// BYO-key sync ask "which assets does this user hold" without market importing
// portfolio and closing the cycle.

// ServiceDeps is the infrastructure the use cases need. No domain module
// appears here — that is the property that lets the service be built first.
type ServiceDeps struct {
	DB      *pgxpool.Pool
	Storage fiber.Storage
	Log     logger.Logger
	// Providers builds a provider chain from a user's own keys. There is no
	// process-wide Provider any more: under BYO-key the application holds no
	// provider credentials of its own.
	Providers marketdata.Factory
	// PublicRates is the keyless feed behind the shared exchange rates. It sits
	// beside Providers rather than inside it because it is the one source that
	// needs no credential, which is also why what it returns may be served to
	// every user instead of only the one who paid for it.
	//
	// Optional: leave it nil and the shared rates stay whatever an admin
	// entered, which is what they were before the feed existed.
	PublicRates marketdata.PublicRateSource
	// Keyring seals and opens those keys.
	Keyring *secretbox.Keyring
}

// Deps is the routing half: the service NewService returned, plus the guards
// and the holdings reader, which only exist once the other modules are built.
type Deps struct {
	Service        *service
	AuthMiddleware authMiddleware
	Limiter        fiber.Handler
	// Holdings answers "which assets does this user own", satisfied by the
	// portfolio module.
	Holdings Holdings
	// CredentialLimiter gates the credential routes more tightly than Limiter:
	// each save and each verification spends the user's own provider quota.
	CredentialLimiter fiber.Handler
}

type Module struct {
	service     *service
	authMiddl   authMiddleware
	handler     *handler
	limiter     fiber.Handler
	credLimiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
}

// NewService builds the module's use cases. It is constructed before portfolio,
// which consumes it.
func NewService(deps ServiceDeps) *service {
	return newService(NewPostgresRepository(deps.DB), deps.Storage, deps.Providers, deps.PublicRates, deps.Keyring, deps.Log)
}

// New completes the module with its HTTP surface. deps.Service must be the
// value NewService returned, so portfolio and these routes share one service.
//
// A missing service or guard panics here rather than at the first request:
// both are wiring, so the composition root is the only thing that can get them
// wrong, and failing at boot is what keeps a misconfigured build from reaching
// production quietly.
func New(deps Deps) *Module {
	if deps.Service == nil {
		panic("market.New: Deps.Service is required — pass the value NewService returned")
	}
	if deps.AuthMiddleware == nil {
		panic("market.New: Deps.AuthMiddleware is required — every market route is guarded by it")
	}

	return new(Module{
		service:     deps.Service,
		handler:     new(handler{deps.Service, deps.Holdings}),
		authMiddl:   deps.AuthMiddleware,
		limiter:     httpx.OrPassThrough(deps.Limiter),
		credLimiter: httpx.OrPassThrough(deps.CredentialLimiter),
	})
}

func (m *Module) Service() *service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	assests := router.Group("/assets")

	assests.Use(m.authMiddl.RequireAuth(), m.limiter)

	admin := httpx.RequireAdmin()

	// Open to every user. An admin's call curates the row for everybody; anybody
	// else's contributes one visible to them alone, capped per day, and never
	// overwriting an existing ticker. The admin guard that used to sit here
	// bought nothing: the transaction importer has always written to this same
	// table for any user with a spreadsheet, and all the guard did was leave the
	// manual path with no way to add an instrument the catalog was missing.
	assests.Post("", m.handler.CreateAsset)
	// The bulk path stays curated: an uploaded sheet upserts, so it can rewrite
	// rows other users hold. A user's bulk route is the transaction importer.
	assests.Post("/import", admin, m.handler.ImportAssets)
	// Editing is admin-only for the same reason the bulk path is: this one
	// addresses an existing row by id, so it rewrites what every user holding
	// that asset sees. Registered after /import so the literal segment is not
	// swallowed by the parameter.
	assests.Patch("/:id", admin, m.handler.UpdateAsset)

	exchangeRates := router.Group("/exchange-rates")
	exchangeRates.Use(m.authMiddl.RequireAuth(), m.limiter)

	exchangeRates.Get("", admin, paginate.New(), m.handler.GetExchangeRates)
	// Open to every signed-in user: the shared rates are what their own figures
	// are converted with, and the feed that fills them is public data fetched
	// with no key. Registered before the parametric PATCH below only by
	// convention — the methods differ, so Fiber cannot confuse the two.
	exchangeRates.Get("/latest", m.handler.GetLatestExchangeRates)
	exchangeRates.Post("", admin, m.handler.CreateExchangeRate)
	exchangeRates.Post("/import", admin, m.handler.ImportExchangeRates)
	// Admin-only and outbound, so it carries the credential limiter rather than
	// the group's: it is the one route here that makes a request to a third
	// party, and the tighter gate is what stops it being used to hammer one.
	exchangeRates.Post("/refresh", admin, m.credLimiter, m.handler.RefreshExchangeRates)
	exchangeRates.Patch("/:id", admin, m.handler.UpdateExchangeRate)

	// BYO-key. Every route here acts on the caller's own keys and holdings —
	// the user id comes from the auth locals, never from the path — so there is
	// no admin variant and no way to name somebody else's credential.
	//
	// The credential routes carry a much tighter limiter than the group's:
	// saving and verifying both spend the user's own provider quota.
	marketData := router.Group("/market")
	marketData.Use(m.authMiddl.RequireAuth(), m.limiter)

	marketData.Get("/credentials", m.handler.ListCredentials)
	marketData.Put("/credentials/:provider", m.credLimiter, m.handler.SaveCredential)
	marketData.Post("/credentials/:provider/verify", m.credLimiter, m.handler.VerifyCredential)
	marketData.Delete("/credentials/:provider", m.credLimiter, m.handler.DeleteCredential)

	marketData.Post("/sync", m.credLimiter, m.handler.SyncMarketData)
	// One asset, one named key. Same gate as the sync above, and for the same
	// reason: both spend the caller's own provider quota.
	marketData.Post("/assets/:assetId/refresh", m.credLimiter, m.handler.RefreshAssetPrice)
}
