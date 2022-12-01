package domain

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

type Base struct {
	ID        string          `json:"id,omitempty" param:"id"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	DeletedAt *mysql.NullTime `json:"deleted_at,omitempty"`
}
