package auth

import (
	"github.com/epsierra/phinex-blog-api/src/models"
)

type GetTokenDto struct {
	Email string `json:"email" validate:"required,email"`
}

type TokenResponse struct {
	Message string      `json:"message"`
	Token   string      `json:"token"`
	User    models.User `json:"user"`
}

type LoginDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}