package domain

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

type Base struct {
	ID        string         `json:"id"`
	UpdatedAt time.Time      `json:"updated_at" `
	CreatedAt time.Time      `json:"created_at" `
	DeletedAt mysql.NullTime `json:"deleted_at" `
}
