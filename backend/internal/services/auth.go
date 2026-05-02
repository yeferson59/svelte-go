package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yeferson59/svelte-go/internal/dtos/auth"
	"github.com/yeferson59/svelte-go/pkg/helpers"
	"golang.org/x/crypto/bcrypt"
)

func (s *Services) Login(ctx context.Context, email, password string) (auth.LoginResponseDTO, error) {
	user, account, err := s.repos.Login(ctx, email)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if !account.ComparePassword(password) {
		return auth.LoginResponseDTO{}, errors.New("invalid credentials")
	}

	expiresAt := time.Now().Add(s.cfg.JWTDuration)

	jwToken, err := s.CreateJWToken(user.ID, user.Email, expiresAt)
	if err != nil {
		return auth.LoginResponseDTO{}, err
	}

	if err := s.repos.CreateSession(ctx, user.ID, jwToken, expiresAt); err != nil {
		return auth.LoginResponseDTO{}, err
	}

	return auth.LoginResponseDTO{
		ID:          user.ID,
		Email:       user.Email,
		AccessToken: jwToken,
	}, nil
}

func (s *Services) CreateJWToken(userID uuid.UUID, email string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"id":    userID,
		"email": email,
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Services) Register(ctx context.Context, name, email, password string) (auth.RegisterResponseDTO, error) {
	_, err := s.repos.GetUserByEmail(ctx, email)
	if err == nil {
		return auth.RegisterResponseDTO{}, errors.New("user existing")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

	user, err := s.repos.Register(ctx, helpers.NormalizateNames(name), email, string(passwordHash))
	if err != nil {
		return auth.RegisterResponseDTO{}, err
	}

	return auth.RegisterResponseDTO{
		Name:  user.Name,
		Email: user.Email,
		Image: user.Image,
	}, nil
}
