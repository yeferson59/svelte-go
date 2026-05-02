package auth

import "github.com/google/uuid"

type RegisterResponseDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Image string `json:"image"`
}

type LoginResponseDTO struct {
	ID          uuid.UUID `json:"name"`
	Email       string    `json:"email"`
	AccessToken string    `json:"accessToken"`
}
