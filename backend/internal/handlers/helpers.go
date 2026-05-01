package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (handler *Handlers) getParamUUID(c fiber.Ctx, paramName string) (uuid.UUID, error) {
	id := c.Params(paramName)
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return idUUID, nil
}

func (handler *Handlers) GetParamID(c fiber.Ctx, paramName string) (string, error) {
	id := c.Params(paramName)

	return id, nil
}

func (handler *Handlers) responseStatusOk(c fiber.Ctx, message, details string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"data":      data,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseSuccess(c fiber.Ctx, status int, message, details string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success":   true,
		"message":   message,
		"details":   details,
		"data":      data,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseBadRequest(c fiber.Ctx, message, details string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseInternalServerError(c fiber.Ctx, message, details string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"details":   details,
		"timestamp": time.Now(),
	})
}

func (handler *Handlers) responseFromDomain(c fiber.Ctx, err error, message, action string) error {
	if strings.Contains(err.Error(), "failed") || strings.Contains(err.Error(), "invalid") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"message":   message,
			"action":    action,
			"timestamp": time.Now(),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success":   false,
		"message":   message,
		"action":    action,
		"timestamp": time.Now(),
	})
}
