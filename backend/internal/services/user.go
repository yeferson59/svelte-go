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
