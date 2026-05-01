package user

type CreateDTO struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email,min=2"`
	Image string `json:"image,omitzero"`
}

type UpdateDTO struct {
	Name  string `json:"name,omitzero" validate:"min=2"`
	Email string `json:"email,omitzero" validate:"min=2"`
	Image string `json:"image,omitzero" validate:"min=2"`
}
