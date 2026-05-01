package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/dtos/user"
	"github.com/yeferson59/svelte-go/internal/entities"
	"github.com/yeferson59/svelte-go/pkg/dtos"
)

func (handler *Handlers) GetListUsers(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "paginate info not found"})
	}

	users, err := handler.services.GetListUsers(handler.ctx, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.FilterPagination[[]entities.User, fiber.Map]{
			Data: users,
			MetaData: fiber.Map{
				"page":   paginateInfo.Page,
				"limit":  paginateInfo.Limit,
				"offset": paginateInfo.Offset,
			},
		},
	)
}

func (handler *Handlers) GetUserByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	user, err := handler.services.GetUserByID(handler.ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (handler *Handlers) CreateUser(c fiber.Ctx) error {
	var createUserDto user.CreateDTO

	if err := c.Bind().Body(&createUserDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": err.Error()})
	}

	user, err := handler.services.CreateUser(handler.ctx, createUserDto.Name, createUserDto.Email, createUserDto.Image)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (handler *Handlers) UpdateUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var updateUser user.UpdateDTO

	if err := c.Bind().Body(&updateUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := handler.services.UpdateUser(handler.ctx, userID, updateUser.Name, updateUser.Email, updateUser.Image)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (handler *Handlers) DeleteUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := handler.services.DeleteUser(handler.ctx, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
