package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/yeferson59/svelte-go/internal/dtos/user"
	"github.com/yeferson59/svelte-go/internal/entities"
	"github.com/yeferson59/svelte-go/pkg/dtos"
	"github.com/yeferson59/svelte-go/pkg/helpers"
)

func (handler *Handlers) GetListUsers(c fiber.Ctx) error {
	paginateInfo, ok := paginate.FromContext(c)
	if !ok {
		return handler.responseInternalServerError(c, "", "paginate info not found")
	}

	users, count, err := handler.services.GetListUsers(handler.ctx, uint(paginateInfo.Offset), uint(paginateInfo.Limit))
	if err != nil {
		return handler.responseFromDomain(c, err, "get product pagination", "users:list")
	}

	totalPages := helpers.CalculateTotalPages(count, uint(paginateInfo.Limit))

	return handler.responseStatusOk(c, "product pagination", "get products successfully", dtos.FilterPagination[[]entities.User, fiber.Map]{
		Items: users,
		MetaData: fiber.Map{
			"currentPage":  paginateInfo.Page,
			"usersForPage": paginateInfo.Limit,
			"offset":       paginateInfo.Offset,
			"totalUsers":   count,
			"totalPages":   totalPages,
			"previous":     paginateInfo.Page > 1,
			"next":         paginateInfo.Page < int(totalPages),
		},
	})
}

func (handler *Handlers) GetUserByID(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "validate id", "invalid user id")
	}

	user, err := handler.services.GetUserByID(handler.ctx, userID)
	if err != nil {
		return handler.responseFromDomain(c, err, "get user by id", "users:id")
	}

	return handler.responseStatusOk(c, "get user by id", "get user successfully", user)
}

func (handler *Handlers) CreateUser(c fiber.Ctx) error {
	var createUserDto user.CreateDTO

	if err := c.Bind().Body(&createUserDto); err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	user, err := handler.services.CreateUser(handler.ctx, createUserDto.Name, createUserDto.Email, createUserDto.Image)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "users:create")
	}

	return handler.responseSuccess(c, fiber.StatusCreated, "", "", user)
}

func (handler *Handlers) UpdateUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	var updateUser user.UpdateDTO

	if err := c.Bind().Body(&updateUser); err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	user, err := handler.services.UpdateUser(handler.ctx, userID, updateUser.Name, updateUser.Email, updateUser.Image)
	if err != nil {
		return handler.responseFromDomain(c, err, "", "users:update")
	}

	return handler.responseStatusOk(c, "", "", user)
}

func (handler *Handlers) DeleteUser(c fiber.Ctx) error {
	userID, err := handler.getParamUUID(c, "id")
	if err != nil {
		return handler.responseBadRequest(c, "", err.Error())
	}

	if err := handler.services.DeleteUser(handler.ctx, userID); err != nil {
		return handler.responseFromDomain(c, err, "", "users:delete")
	}

	return handler.responseSuccess(c, fiber.StatusNoContent, "", "", "")
}
