package repositories

import (
	"context"
	"errors"

	"github.com/yeferson59/svelte-go/internal/entities"
)

func (r *Repository) Login(ctx context.Context, email, password string) {
	r.db.QueryRow(ctx, "SELECT u.*, a.* FROM users u JOIN accounts a ON u.id = a.user_id WHERE u.email = $1", email).Scan()
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
