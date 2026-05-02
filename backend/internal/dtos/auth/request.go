package auth

type RegisterRequestDTO struct {
	Name     string `json:"name" validate:"required,min=2,max=254"`
	Email    string `json:"email" validate:"required,email,min=2,max=254"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}
