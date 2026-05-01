package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/entities"
)

func (s *Services) GetListUsers(ctx context.Context, offset, limit uint) ([]entities.User, error) {
	return s.repos.ListUsers(ctx, offset, limit)
}

func (s *Services) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return s.repos.GetUserByID(ctx, id)
}

func (s *Services) CreateUser(ctx context.Context, name, email, image string) (entities.User, error) {
	return s.repos.CreateUser(ctx, name, email, image)
}

func (s *Services) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	return s.repos.UpdateUser(ctx, id, name, email, image)
}

func (s *Services) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repos.DeleteUser(ctx, id)
}
