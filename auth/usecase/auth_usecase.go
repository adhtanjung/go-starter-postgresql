package usecase

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/pkg/helpers"
	"github.com/adhtanjung/go-boilerplate/user/usecase/helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type authUsecase struct {
	userRepo       domain.UserRepository
	userRoleRepo   domain.UserRoleRepository
	roleRepo       domain.RoleRepository
	contextTimeout time.Duration
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

func NewAuthUsecase(u domain.UserRepository, ur domain.UserRoleRepository, r domain.RoleRepository, timeout time.Duration) domain.AuthUsecase {
	return &authUsecase{
		userRepo:       u,
		userRoleRepo:   ur,
		roleRepo:       r,
		contextTimeout: timeout,
	}
}
func (a *authUsecase) Login(c context.Context, auth domain.Auth) (string, string, error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	user, err := a.userRepo.GetOneByUsernameOrEmail(ctx, auth.UsernameOrEmail)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}

	if match := helper.CheckPasswordHash(auth.Password, user.Password); !match {
		return "", "", &AuthError{"incorrect username or password"}
	}
	token, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 24,
		Secret:    "",
	})
	if err != nil {
		return "", "", err
	}
	refreshToken, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 72,
		Secret:    viper.GetString("secret.refresh_jwt"),
	})
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func (u *authUsecase) Register(c context.Context, m *domain.User, ur *domain.UserRole) (err error) {

	var emptyUser domain.User
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	queryUsername := domain.UserQueryArgs{
		WhereClause: domain.WhereClause{
			User: domain.Query{
				Args:   m.Username,
				Clause: "username = ?",
			},
		},
	}

	isUsernameTaken, err := u.userRepo.GetOne(ctx, queryUsername)
	if err != nil {
		logrus.Error("fetch username failed, error: ", err.Error())
		return
	}
	if !reflect.DeepEqual(isUsernameTaken, emptyUser) {
		err = errors.New("username already taken")
		return
	}
	queryEmail := domain.UserQueryArgs{WhereClause: domain.WhereClause{User: domain.Query{Args: m.Email, Clause: "email = ?"}}}
	isEmailTaken, err := u.userRepo.GetOne(ctx, queryEmail)
	if err != nil {
		fmt.Printf("fetch user email failed, error: '%s'", err.Error())
		return
	}
	if !reflect.DeepEqual(isEmailTaken, emptyUser) {
		err = errors.New("email already taken")
		return
	}
	hashed, err := helper.HashPassword(m.Password)
	if err != nil {
		fmt.Printf("password hashing failed, error: '%s'", err.Error())
	}

	defaultRole, err := u.roleRepo.GetByName(ctx, "user")
	if err != nil {
		fmt.Printf("fetch default role failed, error: '%s'", err.Error())
		return
	}

	now := time.Now()
	// m.Role = defaultRole
	m.CreatedAt = &now
	m.UpdatedAt = &now
	m.Password = hashed
	m.IsVerified = false

	err = u.userRepo.Store(ctx, m)
	ur.CreatedAt = &now
	ur.UpdatedAt = &now
	ur.RoleID = defaultRole.ID
	ur.UserID = m.ID
	err = u.userRoleRepo.Store(ctx, ur)
	if err != nil {
		return
	}

	user, err := u.userRepo.GetOneByUsernameOrEmail(ctx, m.Email)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}
	token, err := helpers.GenerateToken(m.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 24,
		Secret:    "",
	})
	if err != nil {
		return
	}
	template, err := ioutil.ReadFile("./web/email_verif.html")
	if err != nil {
		return
	}
	data := struct {
		Token string
	}{
		Token: token,
	}
	err = helpers.SendEmail(template, data, m.Email)
	return

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
	token, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{ExpiresAt: 1, Secret: viper.GetString("secret.refresh_jwt")})
	if err != nil {
		return
	}
	b, err := ioutil.ReadFile("./web/reset_pass.html")
	if err != nil {
		panic(err)
	}
	// Define the data that will be used to fill the template
	data := struct {
		ResetPasswordLink string
	}{
		ResetPasswordLink: fmt.Sprintf("https://example.com/reset-password?token=%s", token),
	}
	err = helpers.SendEmail(b, data, user.Email)

	return
}
