package services

import (
	"context"

	"github.com/yeferson59/svelte-go/internal/repositories"
)

type Services struct {
	ctx   context.Context
	repos repositories.Repository
}

func New(ctx context.Context, repos repositories.Repository) Services {
	return Services{
		ctx:   ctx,
		repos: repos,
	}
}
