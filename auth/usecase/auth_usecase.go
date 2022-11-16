package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/user/usecase/helper"
)

type authUsecase struct {
	userRepo       domain.UserRepository
	contextTimeout time.Duration
}
type jwtCustomClaims struct {
	User domain.User `json:"user"`
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

func NewAuthUsecase(u domain.UserRepository, timeout time.Duration) domain.AuthUsecase {
	return &authUsecase{
		userRepo:       u,
		contextTimeout: timeout,
	}
}

func (a *authUsecase) Login(c context.Context, auth domain.Auth) (string, error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	user, err := a.userRepo.GetOneByUsername(ctx, auth.Username)

	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}
	if match := helper.CheckPasswordHash(auth.Password, user.Password); !match {
		return "", &AuthError{"incorrect username or password"}
	}
	// Set custom claims

	claims := &jwtCustomClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	// err = a.authRepo.Login(ctx, m)
	return t, nil
}
