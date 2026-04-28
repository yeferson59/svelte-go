package middlewares

import "github.com/yeferson59/svelte-go/internal/config"

type Middlewares struct {
	envs *config.Env
}

func New(envs *config.Env) Middlewares {
	return Middlewares{
		envs: envs,
	}
}
