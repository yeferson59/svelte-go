package handlers

import (
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
