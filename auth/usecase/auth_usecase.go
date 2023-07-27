package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/adhtanjung/go-starter/domain"
	"github.com/adhtanjung/go-starter/pkg/helpers"
	"github.com/adhtanjung/go-starter/user/usecase/helper"
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
func (a *authUsecase) Login(c context.Context, auth domain.Auth, isOauth bool) (data domain.AuthResponse, err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	user, err := a.userRepo.GetOneByUsernameOrEmail(ctx, auth.UsernameOrEmail)
	if err != nil {
		fmt.Printf("fetching user failed: '%s'", err.Error())
	}

	if !isOauth {
		if match := helper.CheckPasswordHash(auth.Password, user.Password); !match {
			return domain.AuthResponse{}, &AuthError{"incorrect username or password"}
		}
	}

	token, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 24,
		Secret:    viper.GetString("secret.jwt"),
	})
	if err != nil {
		return domain.AuthResponse{}, err
	}
	refreshToken, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 24 * 7,
		Secret:    viper.GetString("secret.refresh_jwt"),
	})
	if err != nil {
		return domain.AuthResponse{}, err
	}
	data = domain.AuthResponse{
		Username:     user.Username,
		Email:        user.Email,
		Gender:       user.Gender,
		Status:       user.Status,
		Token:        token,
		RefreshToken: refreshToken,
	}

	return data, nil

	// return token, refreshToken, nil
}

func (u *authUsecase) Register(c context.Context, m *domain.User, ur *domain.UserRole, isOauth bool) (data domain.AuthResponse, err error) {

	var emptyUser domain.User
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	queryEmail := domain.QueryArgs{
		WhereClause: domain.WhereClause{
			User: domain.Query{
				Args: m.Email, Clause: "email = ?",
			},
		},
	}
	isEmailTaken, err := u.userRepo.GetOne(ctx, queryEmail)
	if err != nil {
		fmt.Printf("fetch user email failed, error: '%s'", err.Error())
		return
	}
	if !reflect.DeepEqual(isEmailTaken, emptyUser) {
		err = errors.New("email already taken")
		return
	}

	if !isOauth {
		queryUsername := domain.QueryArgs{
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
			return domain.AuthResponse{}, err
		}
		if !reflect.DeepEqual(isUsernameTaken, emptyUser) {
			err = errors.New("username already taken")
			return domain.AuthResponse{}, err
		}
		hashed, err := helper.HashPassword(m.Password)
		if err != nil {
			fmt.Printf("password hashing failed, error: '%s'", err.Error())
			return domain.AuthResponse{}, err
		}
		m.Password = hashed
	}

	defaultRole, err := u.roleRepo.GetByName(ctx, "user")
	if err != nil {
		fmt.Printf("fetch default role failed, error: '%s'", err.Error())
		return domain.AuthResponse{}, err
	}

	now := time.Now()
	// m.Role = defaultRole
	m.CreatedAt = &now
	m.UpdatedAt = &now

	err = u.userRepo.Store(ctx, m)
	ur.CreatedAt = &now
	ur.UpdatedAt = &now
	ur.RoleID = defaultRole.ID
	ur.UserID = m.ID
	err = u.userRoleRepo.Store(ctx, ur)
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("error user role: %s", err))
		return domain.AuthResponse{}, err
	}
	findByEmail := domain.QueryArgs{
		WhereClause: domain.WhereClause{
			User: domain.Query{
				Args: m.Email, Clause: "email = ?",
			},
		},
	}
	user, err := u.userRepo.GetOne(ctx, findByEmail)
	if err != nil {
		return domain.AuthResponse{}, fmt.Errorf(fmt.Sprintf("error get user: %s", err))
	}
	if !isOauth {
		tokenTemplate, err := helpers.GenerateToken(m.ID, user.UserRoles, helpers.ShouldClaims{
			ExpiresAt: 24,
			Secret:    "",
		})
		if err != nil {
			return domain.AuthResponse{}, fmt.Errorf(fmt.Sprintf("error template: %s", err))
		}
		dir, err := os.Getwd()
		if err != nil {
			return domain.AuthResponse{}, fmt.Errorf("error get wd")
		}
		template, err := os.ReadFile(filepath.Join(dir, "/web/email_verif.html"))
		if err != nil {
			return domain.AuthResponse{}, fmt.Errorf("error read file")
		}
		data := struct {
			Token string
		}{
			Token: tokenTemplate,
		}
		emailChan := make(chan *helpers.Email)

		err = helpers.SendEmail(emailChan)
		emailChan <- &helpers.Email{Recipient: []string{m.Email}, Template: template, Body: data}
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("error sending email: %s", err))
			return domain.AuthResponse{}, err
		}
		// err = helpers.SendEmail(template, data, m.Email)
		// if err != nil {
		// 	return "", "", err
		// }
	}
	token, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 24,
		Secret:    viper.GetString("secret.jwt"),
	})
	if err != nil {
		return domain.AuthResponse{}, errors.New("error buat token")
	}
	refreshToken, err := helpers.GenerateToken(user.ID, user.UserRoles, helpers.ShouldClaims{
		ExpiresAt: 72,
		Secret:    viper.GetString("secret.refresh_jwt"),
	})
	if err != nil {
		return domain.AuthResponse{}, errors.New("error buat refresh token")
	}

	data = domain.AuthResponse{
		Username:     user.Username,
		Email:        user.Email,
		Gender:       user.Gender,
		Status:       user.Status,
		Token:        token,
		RefreshToken: refreshToken,
	}

	return data, nil

}

func (a *authUsecase) ForgotPassword(c context.Context, email string) (err error) {

	emailChan := make(chan *helpers.Email)

	err = helpers.SendEmail(emailChan)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()

	query := domain.QueryArgs{
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
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	b, err := os.ReadFile(filepath.Join(dir, "/web/reset_pass.html"))
	if err != nil {
		return
	}
	// Define the data that will be used to fill the template
	data := struct {
		ResetPasswordLink string
	}{
		ResetPasswordLink: fmt.Sprintf("https://example.com/reset-password?token=%s", token),
	}
	emailChan <- &helpers.Email{Recipient: []string{user.Email}, Template: b, Body: data}
	// time.Sleep(3 * time.Second)
	// err = helpers.SendEmail(b, data, user.Email)

	return
}
