package domain

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
)

type User struct {
	Base
	Username   string                `json:"username,omitempty" validate:"required" form:"username" gorm:"index"`
	Email      string                `json:"email,omitempty" validate:"required" form:"email" gorm:"index"`
	Password   string                `json:"password,omitempty" validate:"required" form:"password"`
	Name       string                `json:"name,omitempty" form:"name"`
	UserRoles  []UserRole            `gorm:"foreignKey:UserID;" json:"user_role,omitempty" form:"user_roles"`
	ProfilePic string                `json:"profile_pic,omitempty"`
	File       *multipart.FileHeader `gorm:"-" json:"file,omitempty"`
	IsVerified bool                  `json:"is_verified,omitempty" form:"is_verified"`
}

type UserUpdate struct {
	Base
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty" validate:"email"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}
type Query struct {
	Args   string
	Clause string
}
type UserQueryArgs struct {
	SelectClause
	WhereClause
}
type SelectClause struct {
	User      string
	UserRoles string
	Role      string
}
type WhereClause struct {
	UserRoles Query
	Role      Query
	User      Query
}

type Auth struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
}
type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

type AuthUsecase interface {
	Login(ctx context.Context, auth Auth) (string, string, error)
	Register(context.Context, *User, *UserRole) error
	ForgotPassword(ctx context.Context, email string) error
}
type UserUsecase interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Store(context.Context, *User, *UserRole) error
	Update(ctx context.Context, a *User) error
	ResendEmailVerification(ctx context.Context, token string) error
}

type UserRepository interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetOne(ctx context.Context, args UserQueryArgs) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Store(ctx context.Context, a *User) error
	Update(ctx context.Context, a *User) error
}

type UserFilepath struct {
	Base
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
	Path     string `json:"path"`
	UserID   uuid.UUID
	User     User `json:"user"`
}

type UserFilepathRepository interface {
	Store(ctx context.Context, f *UserFilepath) error
	// GetByUserID(ctx context.Context, userID string) ([]UserFilepath, error)
}

// func (u *User) BeforeSave(tx *gorm.DB) error {
// 	hashedPassword, err := helper.HashPassword(u.Password)
// 	if err != nil {
// 		return err
// 	}
// 	u.Password = string(hashedPassword)
// 	return nil
// }
