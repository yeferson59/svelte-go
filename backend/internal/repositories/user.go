package repositories

import (
	"context"

	"github.com/yeferson59/svelte-go/internal/entities"
)

func (r *Repository) List(ctx context.Context, offset, limit uint) ([]entities.User, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM users LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User

	for rows.Next() {
		var user entities.User

		if err := rows.Scan(&user); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
