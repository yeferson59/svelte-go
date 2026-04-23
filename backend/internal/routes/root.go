package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/svelte-go/internal/handlers"
)

type Routes struct {
	app      *fiber.App
	handlers handlers.Handlers
}

func New(app *fiber.App, handlers handlers.Handlers) Routes {
	return Routes{
		app:      app,
		handlers: handlers,
	}
}

func (r Routes) Init() {
	r.Health()
}
