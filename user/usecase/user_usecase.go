package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/bxcodec/go-clean-arch/user/usecase/helper"
)

type userUseCase struct {
	userRepo       domain.UserRepository
	contextTimeout time.Duration
}

func NewUserUsecase(u domain.UserRepository, timeout time.Duration) domain.UseruUsecase {
	return &userUseCase{
		userRepo:       u,
		contextTimeout: timeout,
	}

}

// func (u *userUseCase) GetByUsername(c context.Context, title string) (res domain.User, err error) {
// 	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
// 	defer cancel()
// 	res, err = u.userRepo.GetByUsername(ctx, title)
// 	if err != nil {
// 		return
// 	}

// 	resAuthor, err := a.authorRepo.GetByID(ctx, res.Author.ID)
// 	if err != nil {
// 		return domain.Article{}, err
// 	}

// 	res.Author = resAuthor
// 	return
// }
func (u *userUseCase) Store(c context.Context, m *domain.User) (err error) {

	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()

	hashed, err := helper.HashPassword(m.Password)
	if err != nil {

		fmt.Printf("password hashing failed, error: '%s'", err.Error())
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.Password = hashed

	err = u.userRepo.Store(ctx, m)
	return
}

func (u *userUseCase) GetOneByUsername(c context.Context, username string) (res domain.User, err error) {
	ctx, cancel := context.WithTimeout(c, u.contextTimeout)
	defer cancel()
	res, err = u.userRepo.GetOneByUsername(ctx, username)
	if err != nil {
		return
	}
	return
}
