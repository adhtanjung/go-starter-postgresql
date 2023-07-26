package domain

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Base
	Username      string                `json:"username,omitempty" form:"username" gorm:"size:191"`
	Email         string                `json:"email,omitempty" validate:"required" form:"email" gorm:"index"`
	Password      string                `json:"password,omitempty" form:"password"`
	Name          string                `json:"name,omitempty" form:"name"`
	Gender        string                `json:"gender,omitempty" form:"gender"`
	Status        string                `gorm:"default:'not active'" json:"status,omitempty" form:"status"`
	UserRoles     []UserRole            `gorm:"foreignKey:UserID;" json:"user_role,omitempty" form:"user_roles"`
	ProfilePic    string                `json:"profile_pic,omitempty"`
	File          *multipart.FileHeader `gorm:"-" json:"file,omitempty"`
	VerifiedAt    *time.Time            `json:"verified_at,omitempty" form:"verified_at"`
	OauthProvider string                `json:"oauth_provider,omitempty"`
	OauthToken    string                `json:"oauth_token,omitempty"`
}

type AuthResponse struct {
	Username     string `json:"username,omitempty" `
	Email        string `json:"email,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Status       string `json:"status,omitempty"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type UserUpdate struct {
	Base
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty" validate:"email"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}
type UserQueryArgs struct {
	SelectClause
	WhereClause
}

type Auth struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
}
type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
}

type AuthUsecase interface {
	Login(ctx context.Context, auth Auth, isOauth bool) (AuthResponse, error)
	Register(context.Context, *User, *UserRole, bool) (AuthResponse, error)
	ForgotPassword(ctx context.Context, email string) error
}
type UserUsecase interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Store(context.Context, *User, *UserRole) error
	Update(ctx context.Context, a *User) error
	ResendEmailVerification(ctx context.Context, token string) error
	GetUsingRefreshToken(ctx context.Context, userID uuid.UUID) (refreshToken string, accessToken string, err error)
}

type UserRepository interface {
	GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (User, error)
	GetOne(ctx context.Context, args QueryArgs) (User, error)
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
