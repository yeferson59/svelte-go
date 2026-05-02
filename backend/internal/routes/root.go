package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/svelte-go/internal/handlers"
	"github.com/yeferson59/svelte-go/internal/middlewares"
)

type Routes struct {
	app         *fiber.App
	router      fiber.Router
	middlewares middlewares.Middlewares
	handlers    handlers.Handlers
}

func New(app *fiber.App, middlewares middlewares.Middlewares, handlers handlers.Handlers) *Routes {
	return new(Routes{
		app:         app,
		middlewares: middlewares,
		handlers:    handlers,
	})
}

func (r *Routes) Init() {
	r.app.Use(
		r.middlewares.Recovery(),
		r.middlewares.RequestID(),
		r.middlewares.Logger(),
		r.middlewares.CORS(),
	)

	r.Health()
	r.Auth()

	r.router = r.app.Use(r.middlewares.JWT())
	r.Users()
}
