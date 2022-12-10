package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id,omitempty" param:"id"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (base *Base) BeforeCreate(scope *gorm.DB) (err error) {
	base.ID = uuid.New()
	return
}
