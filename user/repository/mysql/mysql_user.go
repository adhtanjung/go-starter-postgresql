package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/domain/helper"
	"github.com/casbin/casbin/v2"
	"github.com/sirupsen/logrus"
)

type mysqlUserRepository struct {
	Conn   *sql.DB
	Casbin *casbin.Enforcer
}

// Usercol returns a reference for a column of a User
func UserCol(colname string, u *domain.User) interface{} {
	switch colname {
	case "id":
		return &u.ID
	case "username":
		return &u.Username
	case "email":
		return &u.Email
	case "password":
		return &u.Password
	case "name":
		return &u.Name
	case "created_at":
		return &u.CreatedAt
	case "updated_at":
		return &u.UpdatedAt
	case "deleted_at":
		return &u.DeletedAt
	case "roles_id":
		return &u.Roles[0].Role.ID
	case "roles_name":
		return &u.Roles[0].Role.Name
	case "roles_created_at":
		return &u.Roles[0].Role.CreatedAt
	case "roles_updated_at":
		return &u.Roles[0].Role.UpdatedAt
	case "roles_deleted_at":
		return &u.Roles[0].Role.DeletedAt
	default:
		panic("unknown column " + colname)
	}
}

func NewMysqlUserRepository(Conn *sql.DB, Casbin *casbin.Enforcer) domain.UserRepository {
	return &mysqlUserRepository{Conn, Casbin}
}

func (m *mysqlUserRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.User, err error) {
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

	result = make([]domain.User, 0)
	for rows.Next() {
		t := domain.User{}
		cols := make([]interface{}, colNum)
		for i := 0; i < colNum; i++ {
			cols[i] = UserCol(columns[i], &t)
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

func (m *mysqlUserRepository) Store(ctx context.Context, u *domain.User) (err error) {
	uuid := helper.GenerateUUID()
	query := `INSERT user SET id =?, username=?, email=?, password=?, name=?, updated_at=?, created_at=?, deleted_at=?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	_, err = stmt.ExecContext(ctx, uuid, u.Username, u.Email, u.Password, u.Name, u.UpdatedAt, u.CreatedAt, nil)
	if err != nil {
		return
	}
	u.ID = uuid
	return
}

func (m *mysqlUserRepository) Update(ctx context.Context, u *domain.User) (err error) {
	query := `UPDATE user SET username= ?, email = ?, name= ?, updated_at= ? WHERE ID = ?`
	stmt, err := m.Conn.PrepareContext(ctx, query)

	if err != nil {
		return
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, u.Username, u.Email, u.Name, u.UpdatedAt, u.ID)
	if err != nil {
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affected != 1 {
		err = fmt.Errorf("weird behavior. total affected: %d", affected)
		return
	}

	return

}

func (m *mysqlUserRepository) GetByID(ctx context.Context, id string) (res domain.User, err error) {
	query := `SELECT id, username, email, name FROM user WHERE id = ?`
	list, err := m.fetch(ctx, query, id)
	if err != nil {
		return domain.User{}, err
	}
	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}
	return
}

func (m *mysqlUserRepository) GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (res domain.User, err error) {
	query := `SELECT user.id, user.username, user.email, user.password, user.name FROM user WHERE user.username = ? OR user.email = ?`
	list, err := m.fetch(ctx, query, usernameOrEmail, usernameOrEmail)
	if err != nil {
		return domain.User{}, err
	}
	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}
	return
}

func (m *mysqlUserRepository) GetOne(ctx context.Context, query string, args ...any) (res domain.User, err error) {
	limitOne := ` LIMIT 1`
	list, err := m.fetch(ctx, query+limitOne, args...)
	if err != nil {
		return domain.User{}, err
	}
	if len(list) > 0 {
		res = list[0]
	}
	return
}
