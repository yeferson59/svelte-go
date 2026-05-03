package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/entities"
	"github.com/yeferson59/svelte-go/pkg/helpers"
)

func (s *Services) GetListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error) {
	return s.repos.ListUsers(ctx, offset, limit)
}

func (s *Services) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return s.repos.GetUserByID(ctx, id)
}

func (s *Services) CreateUser(ctx context.Context, name, email string) (entities.User, error) {
	name = helpers.NormalizateNames(name)

	return s.repos.CreateUser(ctx, name, email)
}

func (s *Services) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	existUser, err := s.repos.GetUserByID(ctx, id)
	if err != nil {
		return entities.User{}, err
	}

	if existUser.DeletedAt != nil {
		return entities.User{}, errors.New("not found user")
	}

	if strings.TrimSpace(name) != "" && existUser.Name != name {
		existUser.Name = helpers.NormalizateNames(name)
	}

	if strings.TrimSpace(email) != "" && existUser.Email != email {
		existUser.Email = email
	}

	if strings.TrimSpace(image) != "" && existUser.Image != image {
		existUser.Image = image
	}

	return s.repos.UpdateUser(ctx, existUser.ID, existUser.Name, existUser.Email, existUser.Image)
}

func (s *Services) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repos.DeleteUser(ctx, id)
}
