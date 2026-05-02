package services

import (
	"github.com/yeferson59/svelte-go/internal/config"
	"github.com/yeferson59/svelte-go/internal/repositories"
)

type Services struct {
	repos repositories.Repository
	cfg   *config.Env
}

func New(repos repositories.Repository, cfg *config.Env) Services {
	return Services{
		repos: repos,
		cfg:   cfg,
	}
}
