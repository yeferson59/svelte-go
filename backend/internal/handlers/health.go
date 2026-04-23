package handlers

import "github.com/gofiber/fiber/v3"

func (h *Handlers) HealthStatus(c fiber.Ctx) {
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
