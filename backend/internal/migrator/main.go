package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/yeferson59/svelte-go/internal/config"
)

func main() {
	cfg := config.New()

	m, err := migrate.New(cfg.LoadEnvs().PathMigration, cfg.LoadEnvs().DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}
