package entities

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt string    `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type Account struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"userId"`
	AccountID             string    `json:"accountId"`
	ProviderID            string    `json:"provider"`
	AccessToken           string    `json:"accessToken"`
	RefreshToken          string    `json:"refreshToken"`
	AccessTokenExpiresAt  string    `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt string    `json:"refreshTokenExpiresAt"`
	Scope                 string    `json:"scope"`
	IDToken               string    `json:"idToken"`
	Password              string    `json:"password"`
	CreatedAt             string    `json:"createdAt"`
	UpdatedAt             string    `json:"updatedAt"`
}

func (a *Account) ComparePassword(password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password)) != nil {
		return false
	}

	return true
}

type Verification struct {
	ID         uuid.UUID `json:"id"`
	Identifier string    `json:"identifier"`
	Value      string    `json:"value"`
	ExpiresAt  string    `json:"expiresAt"`
	CreatedAt  string    `json:"createdAt"`
	UpdatedAt  string    `json:"updatedAt"`
}
