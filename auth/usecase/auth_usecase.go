package usecase

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/pkg/helpers"
	"github.com/adhtanjung/go-boilerplate/user/usecase/helper"
)

type authUsecase struct {
	userRepo       domain.UserRepository
	userRoleRepo   domain.UserRoleRepository
	contextTimeout time.Duration
}
type JwtCustomClaims struct {
	UserID uuid.UUID         `json:"id"`
	Roles  []domain.UserRole `json:"roles"`
	// User domain.User `json:"data"`
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
func GenerateToken(userID uuid.UUID, userRoles []domain.UserRole) (generatedToken string, err error) {
	// convert the struct to a map
	m := map[string]interface{}{
		"ID":    userID,
		"Roles": userRoles,
	}
	user := domain.User{
		Base:      domain.Base{ID: m["ID"].(uuid.UUID)},
		UserRoles: m["Roles"].([]domain.UserRole),
	}
	claims := &JwtCustomClaims{
		user.ID,
		user.UserRoles,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	generatedToken, err = token.SignedString([]byte(viper.GetString(`secret.jwt`)))
	if err != nil {
		return "", err
	}
	return
}
func (a *authUsecase) Login(c context.Context, auth domain.Auth) (string, error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	user, err := a.userRepo.GetOneByUsernameOrEmail(ctx, auth.UsernameOrEmail)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}
	// if err != nil {
	// 	// logrus.Error(err.Error())
	// 	fmt.Printf("fetching user role failed: '%s'", err.Error())
	// 	return "", &AuthError{"fetching user role failed"}
	// }

	if match := helper.CheckPasswordHash(auth.Password, user.Password); !match {
		return "", &AuthError{"incorrect username or password"}
	}
	token, err := GenerateToken(user.ID, user.UserRoles)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *authUsecase) ForgotPassword(c context.Context, email string) (err error) {

	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()

	query := domain.UserQueryArgs{
		SelectClause: domain.SelectClause{
			User:      "id, username, email",
			UserRoles: "id, user_id, role_id",
			Role:      "id, name",
		},
		WhereClause: domain.WhereClause{
			User: domain.Query{
				Args:   email,
				Clause: "email = ?",
			},
		}}
	user, err := a.userRepo.GetOne(ctx, query)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}
	token, err := GenerateToken(user.ID, user.UserRoles)
	if err != nil {
		return
	}
	b, err := ioutil.ReadFile("./web/reset_pass.html")
	if err != nil {
		panic(err)
	}
	err = helpers.SendEmail(b, token, user.Email)

	return
}
