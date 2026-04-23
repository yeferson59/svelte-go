package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	Port        string
	DatabaseURL string
}

func (Config) LoadEnvs() *Env {
	_ = godotenv.Load()

	return &Env{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
