package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/svelte-go/internal/entities"
)

func (r *Repository) ListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error) {
	var count uint

	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, "SELECT * FROM users LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]entities.User, 0, limit)

	for rows.Next() {
		var user entities.User

		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, 0, err
		}

		users = append(users, user)
	}

	return users, count, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, name, email string) (entities.User, error) {
	contextTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var user entities.User
	var roleID uuid.UUID

	if tx, err := r.db.BeginTx(contextTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite}); err == nil {
		if err := tx.QueryRow(contextTimeout, "SELECT id FROM roles WHERE name = $1", "customer").Scan(&roleID); err != nil {
			return entities.User{}, errors.New("failed create new user")
		}

		if tx.QueryRow(contextTimeout, "INSERT INTO users (name, email, role_id) VALUES ($1, $2, $3) RETURNING *", name, email, roleID).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt) != nil {
			return entities.User{}, errors.New("failed create new user")
		}

		return user, tx.Commit(contextTimeout)
	}

	return user, errors.New("failed create new user")
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx, "UPDATE users SET name = $1, email = $2, image = $3, updated_at = $4 WHERE id = $5 RETURNING *", name, email, image, time.Now(), id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL", time.Now(), id.String())

	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.UpdatedAt, &user.CreatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}
