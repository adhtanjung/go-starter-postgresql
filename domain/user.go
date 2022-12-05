package domain

import (
	"context"
	"mime/multipart"
)

type User struct {
	Base
	Username   string     `json:"username,omitempty" validate:"required" form:"username"`
	Email      string     `json:"email,omitempty" validate:"required" form:"email"`
	Password   string     `json:"password,omitempty" validate:"required" form:"password"`
	Name       string     `json:"name,omitempty" form:"name"`
	Roles      []UserRole `json:"roles,omitempty" form:"roles"`
	ProfilePic string
	File       *multipart.FileHeader
}

type UserUpdate struct {
	Base
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty" validate:"email"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}

type Auth struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

type AuthUsecase interface {
	Login(ctx context.Context, auth Auth) (string, error)
}
type UserUsecase interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Store(context.Context, *User, *UserRole) error
	Update(ctx context.Context, a *User) error
}

type UserRepository interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetOne(ctx context.Context, query string, args ...any) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Store(ctx context.Context, a *User) error
	Update(ctx context.Context, a *User) error
}

type UserFilepath struct {
	Base
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
	Path     string `json:"path"`
	User     *User  `json:"user"`
}

type UserFilepathRepository interface {
	Store(ctx context.Context, f *UserFilepath) error
	// GetByUserID(ctx context.Context, userID string) ([]UserFilepath, error)
}
