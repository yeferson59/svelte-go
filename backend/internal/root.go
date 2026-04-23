package internal

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/svelte-go/internal/handlers"
	"github.com/yeferson59/svelte-go/internal/repositories"
	"github.com/yeferson59/svelte-go/internal/routes"
	"github.com/yeferson59/svelte-go/internal/services"
)

type Bootstrap struct {
	app *fiber.App
	db  *pgxpool.Pool
}

func New(app *fiber.App, db *pgxpool.Pool) *Bootstrap {
	return new(Bootstrap{
		app: app,
		db:  db,
	})
}

func (b *Bootstrap) Init(ctx context.Context) error {
	repos := repositories.New(ctx, b.db)
	services := services.New(ctx, repos)
	handlers := handlers.New(ctx, services)
	routes := routes.New(b.app, handlers)

	routes.Init()

	return nil
}
