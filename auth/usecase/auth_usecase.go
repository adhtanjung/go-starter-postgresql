package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/user/usecase/helper"
)

type authUsecase struct {
	userRepo       domain.UserRepository
	userRoleRepo   domain.UserRoleRepository
	contextTimeout time.Duration
}
type JwtCustomClaims struct {
	UserID string            `json:"id"`
	Roles  []domain.UserRole `json:"roles"`
	jwt.StandardClaims
}
type CustomError interface {
	Error() string
}

type AuthError struct {
	Message string
}

func (a *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s", a.Message)
}

func NewAuthUsecase(u domain.UserRepository, ur domain.UserRoleRepository, timeout time.Duration) domain.AuthUsecase {
	return &authUsecase{
		userRepo:       u,
		userRoleRepo:   ur,
		contextTimeout: timeout,
	}
}

func (a *authUsecase) Login(c context.Context, auth domain.Auth) (string, error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	user, err := a.userRepo.GetOneByUsernameOrEmail(ctx, auth.UsernameOrEmail)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}
	userRoles, err := a.userRoleRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}

	if match := helper.CheckPasswordHash(auth.Password, user.Password); !match {
		return "", &AuthError{"incorrect username or password"}
	}
	user.Roles = userRoles
	// Set custom claims

	claims := &JwtCustomClaims{
		user.ID,
		user.Roles,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(viper.GetString(`secret.jwt`)))
	if err != nil {
		return "", err
	}

	// err = a.authRepo.Login(ctx, m)
	return t, nil
}
