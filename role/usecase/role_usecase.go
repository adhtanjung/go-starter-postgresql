package usecase

import (
	"context"
	"time"

	"github.com/adhtanjung/go-boilerplate/domain"
)

type roleUsecase struct {
	roleRepo       domain.RoleRepository
	contextTimeout time.Duration
}

func NewRoleUsecase(r domain.RoleRepository, timout time.Duration) domain.RoleUsecase {
	return &roleUsecase{
		roleRepo:       r,
		contextTimeout: timout,
	}
}

func (r *roleUsecase) GetByName(c context.Context, name string) (res domain.Role, err error) {
	ctx, cancel := context.WithTimeout(c, r.contextTimeout)
	defer cancel()
	res, err = r.roleRepo.GetByName(ctx, name)
	if err != nil {
		return
	}
	return

}

func (r *roleUsecase) Store(c context.Context, ro *domain.Role) (err error) {
	ctx, cancel := context.WithTimeout(c, r.contextTimeout)
	defer cancel()
	isRoleExists, _ := r.GetByName(ctx, ro.Name)
	if isRoleExists != (domain.Role{}) {
		return domain.ErrConflict
	}
	now := time.Now()
	ro.CreatedAt = now
	ro.UpdatedAt = now
	err = r.roleRepo.Store(ctx, ro)
	return
}
