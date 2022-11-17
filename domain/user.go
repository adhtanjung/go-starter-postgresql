package domain

import (
	"context"
)

type User struct {
	Base
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" `
	Role     Role   `json:"role"`
}

type Auth struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthUsecase interface {
	Login(ctx context.Context, auth Auth) (string, error)
}
type AuthRepository interface {
	Login(ctx context.Context, username string) error
}
type UseruUsecase interface {
	GetOneByUsername(ctx context.Context, username string) (User, error)
	Store(context.Context, *User) error
}

type UserRepository interface {
	GetOneByUsername(ctx context.Context, username string) (User, error)
	Store(ctx context.Context, a *User) error
}
