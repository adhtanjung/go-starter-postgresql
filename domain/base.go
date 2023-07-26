package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id,omitempty" param:"id"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type History struct {
	Action string `json:"action"`
}

func (base *Base) BeforeCreate(scope *gorm.DB) (err error) {
	base.ID = uuid.New()
	return
}

type Query struct {
	Args, Clause string
}
type QueryArgs struct {
	SelectClause
	WhereClause
}
type SelectClause struct {
	User, UserRoles, Role, Employee string
}
type WhereClause struct {
	UserRoles, Role, User, Employee Query
}
