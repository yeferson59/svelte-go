package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (c *Config) ConnectionDB(ctx context.Context) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, c.LoadEnvs().DatabaseURL)
}
