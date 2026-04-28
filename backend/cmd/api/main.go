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
		log.Fatal("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	if err := internal.New(app, dbPool).Init(ctx); err != nil {
		log.Fatal("failed to initialize app: " + err.Error())
	}

	app.Listen(":"+envs.Port, fiber.ListenConfig{
		EnablePrintRoutes: true,
	})
}
