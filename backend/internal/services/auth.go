package services

import (
	"context"
	"errors"

	"github.com/yeferson59/svelte-go/internal/entities"
	"github.com/yeferson59/svelte-go/pkg/helpers"
)

func (s *Services) Register(ctx context.Context, name, email, password string) (entities.User, error) {
	_, err := s.repos.GetUserByEmail(ctx, email)
	if err == nil {
		return entities.User{}, errors.New("user existing")
	}

	return s.repos.Register(ctx, helpers.NormalizateNames(name), email, password)
}
