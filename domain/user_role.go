package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRole struct {
	Base
	UserID uuid.UUID
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RoleID uuid.UUID
	Role   Role `gorm:"foreignKey:RoleID" json:"role"`
}

// type UserRoleGetByUserIDRes struct {
// 	Role_id   string
// 	Role_name string
// }

type UserRoleRepository interface {
	Store(ctx context.Context, u *UserRole) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
}
