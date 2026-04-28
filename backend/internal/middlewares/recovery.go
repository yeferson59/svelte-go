package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func (Middlewares) Recovery() fiber.Handler {
	return recover.New()
}
