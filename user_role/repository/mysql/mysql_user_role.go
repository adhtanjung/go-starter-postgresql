package mysql

import (
	"context"
	"database/sql"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/domain/helper"
	"github.com/sirupsen/logrus"
)

type mysqlUserRoleRepository struct {
	Conn *sql.DB
}

func NewMysqlUserRoleRepository(Conn *sql.DB) domain.UserRoleRepository {
	return &mysqlUserRoleRepository{Conn}
}

// Usercol returns a reference for a column of a User
func UserRoleCol(colname string, u *domain.UserRole) interface{} {
	switch colname {
	case "role":
		return &u.Role.ID
	case "role_id":
		return &u.Role.ID
	case "role_name":
		return &u.Role.Name
	case "role_created_at":
		return &u.Role.CreatedAt
	case "role_updated_at":
		return &u.Role.UpdatedAt
	case "role_deleted_at":
		return &u.Role.DeletedAt
	default:
		panic("unknown column " + colname)
	}
}

func (m *mysqlUserRoleRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.UserRole, err error) {
	rows, err := m.Conn.QueryContext(ctx, query, args...)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	// get the column names from the query
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		logrus.Error(err)
	}

	colNum := len(columns)

	result = make([]domain.UserRole, 0)
	for rows.Next() {
		t := domain.UserRole{}
		cols := make([]interface{}, colNum)
		for i := 0; i < colNum; i++ {
			cols[i] = UserRoleCol(columns[i], &t)
		}
		err = rows.Scan(
			cols...,
		)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil

}

func (m *mysqlUserRoleRepository) Store(ctx context.Context, r *domain.UserRole) (err error) {
	uuid := helper.GenerateUUID()
	query := `INSERT user_role SET id = ?, user_id = ?, role_id = ?, created_at = ?, updated_at = ?, deleted_at = ?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	_, err = stmt.ExecContext(ctx, uuid, r.User.ID, r.Role.ID, r.CreatedAt, r.UpdatedAt, nil)
	if err != nil {
		return
	}
	r.ID = uuid
	return

}

func (m *mysqlUserRoleRepository) GetByUserID(ctx context.Context, userID string) (res []domain.UserRole, err error) {
	query := "SELECT r.id AS role_id, r.name AS role_name FROM `user_role` as ur INNER JOIN user as u ON u.id = ur.user_id INNER JOIN role as r ON r.id = ur.role_id WHERE ur.user_id = ?"
	list, err := m.fetch(ctx, query, userID)
	if err != nil {
		return []domain.UserRole{}, err
	}
	if len(list) > 0 {
		res = list
	} else {
		return res, domain.ErrNotFound
	}
	return

}
