package main

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/yeferson59/svelte-go/internal"
	"github.com/yeferson59/svelte-go/internal/config"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func main() {
	app, cfg := fiber.New(fiber.Config{
		StructValidator: new(structValidator{validate: validator.New()}),
	}), config.New()
	envs := cfg.LoadEnvs()
	ctx := context.Background()
	dbPool, err := cfg.ConnectionDB(ctx, envs.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	if err := internal.New(app, dbPool, envs).Init(ctx); err != nil {
		log.Fatal("failed to initialize app: " + err.Error())
	}

	app.Listen(":" + envs.Port)
}
