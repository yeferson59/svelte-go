package auth

type RegisterResponseDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Image string `json:"image"`
}

type LoginResponseDTO struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Image       string `json:"image"`
	AccessToken string `json:"accessToken"`
}
