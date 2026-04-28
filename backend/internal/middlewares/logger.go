package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func (Middlewares) Logger() fiber.Handler {
	return logger.New()
}
