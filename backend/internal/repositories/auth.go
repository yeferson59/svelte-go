package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/entities"
)

func (r *Repository) Login(ctx context.Context, email string) (entities.User, entities.Account, error) {
	var account entities.Account
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT u.id, u.email, u.email_verified, a.id, a.provider_id, a.account_id, a.password FROM users u JOIN accounts a ON u.id = a.user_id WHERE u.email = $1 AND u.deleted_at IS NULL", email).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&account.ID,
		&account.ProviderID,
		&account.AccountID,
		&account.Password,
	); err != nil {
		return entities.User{}, entities.Account{}, err
	}

	return user, account, nil
}

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, "INSERT INTO sessions(user_id, token, expires_at) VALUES($1, $2, $3)", userID.String(), token, expiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Register(ctx context.Context, name, email, password string) (entities.User, error) {
	user, err := r.CreateUser(ctx, name, email, "avatar.png")
	if err != nil {
		return entities.User{}, errors.New("error create new user")
	}

	_, err = r.db.Exec(ctx, "INSERT INTO accounts(user_id, account_id, provider_id, password) VALUES($1, $2, $3, $4)", user.ID, "credentials", "local", password)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}
