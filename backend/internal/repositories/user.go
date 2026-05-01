package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/entities"
)

func (r *Repository) ListUsers(ctx context.Context, offset, limit uint) ([]entities.User, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM users LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entities.User, 0, limit)

	for rows.Next() {
		var user entities.User

		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, name, email, image string) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "INSERT INTO users (name, email, image) VALUES ($1, $2, $3) RETURNING *", name, email, image).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx, "UPDATE users SET name = $1, email = $2, image = $3, updated_at = $4 WHERE id = $5 RETURNING *", name, email, image, time.Now(), id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2", time.Now(), id.String())

	return err
}
