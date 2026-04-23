package main

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/yeferson59/svelte-go/internal"
	"github.com/yeferson59/svelte-go/internal/config"
)

func main() {
	app, cfg := fiber.New(), config.New()
	envs := cfg.LoadEnvs()
	ctx := context.Background()
	dbPool, err := cfg.ConnectionDB(ctx)
	if err != nil {
		log.Fatal("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	bootstrap := internal.New(app, dbPool)

	bootstrap.Init(ctx)

	app.Listen(":"+envs.Port, fiber.ListenConfig{
		EnablePrintRoutes: true,
	})
}
