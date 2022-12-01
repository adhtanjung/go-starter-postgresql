package domain

import "context"

type UserRole struct {
	Base
	User *User `json:"user,omitempty"`
	Role Role  `json:"role"`
}

type UserRoleRepository interface {
	Store(ctx context.Context, u *UserRole) error
	GetByUserID(ctx context.Context, userID string) ([]UserRole, error)
}
