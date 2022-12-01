package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/user/usecase/helper"
)

type userUsecase struct {
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	userRoleRepo   domain.UserRoleRepository
	casbinRepo     domain.CasbinRBACRepository
	contextTimeout time.Duration
}

func NewUserUsecase(u domain.UserRepository, r domain.RoleRepository, ur domain.UserRoleRepository, cas domain.CasbinRBACRepository, timeout time.Duration) domain.UserUsecase {
	return &userUsecase{
		userRepo:       u,
		roleRepo:       r,
		userRoleRepo:   ur,
		casbinRepo:     cas,
		contextTimeout: timeout,
	}

}

func (u *userUsecase) Store(c context.Context, m *domain.User, ur *domain.UserRole) (err error) {

	var emptyUser domain.User
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	queryUsername := `SELECT id FROM user WHERE username = ?`
	queryEmail := `SELECT id FROM user WHERE email = ?`
	isUsernameTaken, err := u.userRepo.GetOne(ctx, queryUsername, m.Username)
	if err != nil {
		fmt.Printf("fetch username failed, error: '%s'", err.Error())
		return
	}
	if !reflect.DeepEqual(isUsernameTaken, emptyUser) {
		err = errors.New("username already taken")
		return
	}
	isEmailTaken, err := u.userRepo.GetOne(ctx, queryEmail, m.Email)
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

	err = u.userRepo.Store(ctx, m)
	ur.CreatedAt = &now
	ur.UpdatedAt = &now
	ur.Role = defaultRole
	ur.User = m
	err = u.userRoleRepo.Store(ctx, ur)
	return
}

func (u *userUsecase) GetOneByUsernameOrEmail(c context.Context, usernameOrEmail string) (res domain.User, err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err = u.userRepo.GetOneByUsernameOrEmail(ctx, usernameOrEmail)
	userRole, err := u.userRoleRepo.GetByUserID(ctx, res.ID)
	res.Roles = userRole
	if err != nil {
		return
	}
	return
}

func (u *userUsecase) Update(c context.Context, user *domain.User) (err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err := u.userRepo.GetByID(ctx, user.ID)
	log.Println(res.Username)
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
	// if user.Password != "" {
	// 	hashed, _ := helper.HashPassword(user.Password)
	// 	res.Password = hashed
	// }
	if user.Name != "" {
		res.Name = user.Name
	}
	log.Println(user.Username)
	now := time.Now()
	user = &res
	user.UpdatedAt = &now

	return u.userRepo.Update(ctx, user)

}
func (u *userUsecase) GetByID(c context.Context, id string) (res domain.User, err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err = u.userRepo.GetByID(ctx, id)
	if err != nil {
		return
	}
	return

}
