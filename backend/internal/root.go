package internal

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/svelte-go/internal/config"
	"github.com/yeferson59/svelte-go/internal/handlers"
	"github.com/yeferson59/svelte-go/internal/middlewares"
	"github.com/yeferson59/svelte-go/internal/repositories"
	"github.com/yeferson59/svelte-go/internal/routes"
	"github.com/yeferson59/svelte-go/internal/services"
)

type Bootstrap struct {
	app  *fiber.App
	db   *pgxpool.Pool
	envs *config.Env
}

func New(app *fiber.App, db *pgxpool.Pool, envs *config.Env) *Bootstrap {
	return new(Bootstrap{
		app:  app,
		db:   db,
		envs: envs,
	})
}

func (b *Bootstrap) Init(ctx context.Context) error {
	repos := repositories.New(b.db)
	services := services.New(repos)
	handlers, middlewares := handlers.New(ctx, services), middlewares.New(ctx, b.envs)
	routes := routes.New(b.app, middlewares, handlers)

	routes.Init()

	return nil
}
