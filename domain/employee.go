package domain

import (
	"context"

	"github.com/google/uuid"
)

type Employee struct {
	Base

	NIP         string `json:"NIP" form:"NIP"`
	NIK         string `json:"NIK" form:"NIK"`
	Name        string `json:"name" form:"name"`
	Gender      string `json:"gender" form:"gender"`
	Email       string `json:"email" form:"email"`
	PhoneNumber string `json:"phone_number" form:"phone_number"`
	JoinedAt    string `json:"joined_at" form:"joined_at"`
	JobRole     string `json:"job_role" form:"job_role"`
}

type EmployeeHistory struct {
	Employee
	History
}

type EmployeeRepository interface {
	GetOne(ctx context.Context, args QueryArgs) (Employee, error)
	GetByID(ctx context.Context, id uuid.UUID) (Employee, error)
	Store(ctx context.Context, a *Employee) error
	Update(ctx context.Context, a *Employee) error
}

type EmployeeUsecase interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Store(context.Context, *Employee) error
	Update(ctx context.Context, a *Employee) error
}
