package mysql

import (
	"context"
	"database/sql"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/domain/helper"
	"github.com/sirupsen/logrus"
)

type mysqlRoleRepository struct {
	Conn *sql.DB
}

func NewMysqlRoleRepository(Conn *sql.DB) domain.RoleRepository {
	return &mysqlRoleRepository{Conn}
}

func (m *mysqlRoleRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.Role, err error) {
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

	result = make([]domain.Role, 0)
	for rows.Next() {
		t := domain.Role{}
		err = rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil

}

func (m *mysqlRoleRepository) GetByName(ctx context.Context, name string) (r domain.Role, err error) {
	query := `SELECT * FROM role WHERE name=?`
	list, err := m.fetch(ctx, query, name)
	if err != nil {
		return
	}
	if len(list) > 0 {
		r = list[0]
	} else {
		return r, domain.ErrNotFound
	}
	return
}

func (m *mysqlRoleRepository) Store(ctx context.Context, r *domain.Role) (err error) {

	query := `INSERT role SET id=?, name=?, created_at=?, updated_at=?, deleted_at=?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	uuid := helper.GenerateUUID()
	_, err = stmt.ExecContext(ctx, uuid, r.Name, r.CreatedAt, r.UpdatedAt, nil)
	if err != nil {
		return
	}
	r.ID = uuid
	return
}
