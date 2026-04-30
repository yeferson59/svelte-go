package middlewares

import (
	"context"

	"github.com/yeferson59/svelte-go/internal/config"
)

type Middlewares struct {
	ctx  context.Context
	envs *config.Env
}

func New(ctx context.Context, envs *config.Env) Middlewares {
	return Middlewares{
		ctx:  ctx,
		envs: envs,
	}
}
