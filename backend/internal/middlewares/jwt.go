package middlewares

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
)

func (m *Middlewares) JWT() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.envs.JWTSecret)},
		Extractor:  extractors.FromAuthHeader("Bearer"),
	})
}
