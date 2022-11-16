package domain

import "context"

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RoleUsecase interface {
	GetByName(ctx context.Context, name string) (Role, error)
	Store(ctx context.Context, r *Role) error
}

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (Role, error)
	Store(ctx context.Context, r *Role) error
}
