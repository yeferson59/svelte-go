package middlewares

import (
	"os"

	middleware "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func (Middlewares) Logger() fiber.Handler {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return middleware.New(middleware.Config{
		Logger: &logger,
	})
}
