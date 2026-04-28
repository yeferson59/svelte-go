package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func (Middlewares) CORS() fiber.Handler {
	return cors.New()
}
