package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func New(ctx context.Context, db *pgxpool.Pool) Repository {
	return Repository{
		ctx: ctx,
		db:  db,
	}
}
