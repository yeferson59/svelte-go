package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func (Middlewares) RequestID() fiber.Handler {
	return requestid.New()
}
