package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (c *Config) ConnectionDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, databaseURL)
}
