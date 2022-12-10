package mysql

import (
	"context"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mysqlUserRoleRepository struct {
	Conn *gorm.DB
}

func NewMysqlUserRoleRepository(Conn *gorm.DB) domain.UserRoleRepository {
	return &mysqlUserRoleRepository{Conn}
}

func (m *mysqlUserRoleRepository) Store(ctx context.Context, r *domain.UserRole) (err error) {
	if result := m.Conn.Create(&r); result.Error != nil {
		return result.Error
	}
	return

}

func (m *mysqlUserRoleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (res []domain.UserRole, err error) {
	// var result []domain.UserRoleGetByUserIDRes
	var result []domain.UserRole
	// query := "SELECT ur.id as id, u.id as user_id, ur.role_id as role_id, r.name as role_name FROM user_roles as ur INNER JOIN users as u ON u.id = ur.user_id INNER JOIN roles as r ON r.id = ur.role_id WHERE ur.user_id = ?"
	// scanned := m.Conn.Raw(query, userID).Scan(&result)
	scanned := m.Conn.Preload("Role").Preload("User").Where("user_id = ?", userID).Find(&result)
	// scanned := m.Conn.Table("user_roles").Select("user_roles.id as id, roles.name as role_name").Joins("INNER JOIN users ON users.id = user_id").Joins("INNER JOIN roles ON roles.id = role_id").Where("user_id = ?", userID).Scan(&result)
	// scanned := m.Conn.Preload("User", func(db *gorm.DB) *gorm.DB {
	// 	return db.Select("Username")
	// }).Preload("Role").Where("user_id = ?", userID).Find(&result)
	// log.Println(result)

	if scanned.Error != nil {
		return []domain.UserRole{}, err
	}
	if scanned.RowsAffected > 0 {
		res = result
	} else {
		return res, domain.ErrNotFound
	}
	return

}
