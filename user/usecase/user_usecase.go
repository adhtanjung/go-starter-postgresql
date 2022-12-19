package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/pkg/helpers"
	"github.com/adhtanjung/go-boilerplate/user/usecase/helper"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type userUsecase struct {
	userRepo         domain.UserRepository
	roleRepo         domain.RoleRepository
	userRoleRepo     domain.UserRoleRepository
	casbinRepo       domain.CasbinRBACRepository
	userFilepathRepo domain.UserFilepathRepository
	contextTimeout   time.Duration
}

func NewUserUsecase(u domain.UserRepository, r domain.RoleRepository, ur domain.UserRoleRepository, cas domain.CasbinRBACRepository, uf domain.UserFilepathRepository, timeout time.Duration) domain.UserUsecase {
	return &userUsecase{
		userRepo:         u,
		roleRepo:         r,
		userRoleRepo:     ur,
		casbinRepo:       cas,
		userFilepathRepo: uf,
		contextTimeout:   timeout,
	}

}

func (u *userUsecase) Store(c context.Context, m *domain.User, ur *domain.UserRole) (err error) {

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

func (u *userUsecase) GetOneByUsernameOrEmail(c context.Context, usernameOrEmail string) (res domain.User, err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err = u.userRepo.GetOneByUsernameOrEmail(ctx, usernameOrEmail)
	if err != nil {
		return
	}
	return
}

func (u *userUsecase) ResendEmailVerification(c context.Context, token string) (err error) {

	// template, err := ioutil.ReadFile("./web/email_verif.html")
	// if err != nil {
	// 	return
	// }
	// data := struct {
	// 	Token string
	// }{
	// 	Token: token,
	// }
	// err = helpers.SendEmail(template, data, m.Email)
	return

}

func (u *userUsecase) Update(c context.Context, user *domain.User) (err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return
	}
	// check if a field in user is exists
	if user.Username != "" {
		res.Username = user.Username
	}
	if user.Email != "" {
		res.Email = user.Email
	}
	if user.Password != "" {
		hashed, _ := helper.HashPassword(user.Password)
		res.Password = hashed
	}
	if user.Name != "" {
		res.Name = user.Name
	}
	now := time.Now()
	if user.File != nil {
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		if cond := allowedExts[filepath.Ext(user.File.Filename)]; cond {
			src, err := user.File.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			folderPath := "public/user"
			folder := "user/"
			// Destination
			err = os.MkdirAll(folderPath, os.ModePerm)
			if err != nil {
				return err
			}
			uuid := helpers.GenerateUUID()
			// Destination
			extension := filepath.Ext(user.File.Filename)
			originalName := strings.TrimSuffix(user.File.Filename, extension)
			fileLocation := filepath.Join(folderPath, originalName+"-"+uuid+extension)
			pathToDb := filepath.Join(folder, originalName+"-"+uuid+extension)
			targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer targetFile.Close()

			// Copy
			if _, err = io.Copy(targetFile, src); err != nil {
				return err
			}

			userFilepath := &domain.UserFilepath{Filename: originalName, Mimetype: extension, Path: pathToDb, UserID: res.ID, Base: domain.Base{CreatedAt: &now, UpdatedAt: &now}}

			res.ProfilePic = pathToDb
			err = u.userFilepathRepo.Store(ctx, userFilepath)
			if err != nil {
				return err
			}
		} else {
			return errors.New("invalid extension")
		}

	}

	user = &res
	return u.userRepo.Update(ctx, user)

}

func (u *userUsecase) SendEmailVefification(c context.Context, userID uuid.UUID) (err error) {

	return
}

func (u *userUsecase) GetByID(c context.Context, id uuid.UUID) (res domain.User, err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err = u.userRepo.GetByID(ctx, id)
	if err != nil {
		return
	}
	return

}
