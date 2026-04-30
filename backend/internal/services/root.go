package services

import (
	"github.com/yeferson59/svelte-go/internal/repositories"
)

type Services struct {
	repos repositories.Repository
}

func New(repos repositories.Repository) Services {
	return Services{
		repos: repos,
	}
}
