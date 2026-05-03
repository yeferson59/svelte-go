package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func (m *Middlewares) CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     m.envs.CORSOrigin,
		AllowCredentials: m.envs.CORSEnabled,
	})
}
