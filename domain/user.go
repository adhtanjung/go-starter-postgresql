package domain

import (
	"context"
	"time"

	"github.com/go-sql-driver/mysql"
)

type User struct {
	ID        string         `json:"id"`
	Username  string         `json:"username" validate:"required"`
	Password  string         `json:"password" validate:"required"`
	Name      string         `json:"name" `
	UpdatedAt time.Time      `json:"updated_at" `
	CreatedAt time.Time      `json:"created_at" `
	DeletedAt mysql.NullTime `json:"deleted_at" `
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
	// Fetch(ctx context.Context, cursor string, num int64) ([]User, string, error)
	// GetByID(ctx context.Context, id int64) (User, error)
	// Update(ctx context.Context, ar *User) error
	GetOneByUsername(ctx context.Context, username string) (User, error)
	Store(context.Context, *User) error
	// Delete(ctx context.Context, id int64) error
}

type UserRepository interface {
	// Fetch(ctx context.Context, cursor string, num int64) (res []User, nextCursor string, err error)
	// GetByID(ctx context.Context, id int64) (User, error)
	// Update(ctx context.Context, ar *User) error
	GetOneByUsername(ctx context.Context, username string) (User, error)
	Store(ctx context.Context, a *User) error
	// Delete(ctx context.Context, id int64) error
}
