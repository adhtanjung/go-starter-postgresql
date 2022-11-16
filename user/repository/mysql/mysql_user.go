package mysql

import (
	"context"
	"database/sql"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type mysqlUserRepository struct {
	Conn *sql.DB
}

func NewMysqlUserRepository(Conn *sql.DB) domain.UserRepository {
	return &mysqlUserRepository{Conn}
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

	result = make([]domain.User, 0)
	for rows.Next() {
		t := domain.User{}
		err = rows.Scan(
			&t.ID,
			&t.Username,
			&t.Password,
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

func (m *mysqlUserRepository) Store(ctx context.Context, u *domain.User) (err error) {
	uuid := GenerateUUID()
	query := `INSERT user SET id =?, username=? , password=? , name=? , updated_at=? , created_at= ? , deleted_at=?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	_, err = stmt.ExecContext(ctx, uuid, u.Username, u.Password, u.Name, u.UpdatedAt, u.CreatedAt, nil)
	if err != nil {
		return
	}
	u.ID = uuid
	return
}

func (m *mysqlUserRepository) GetOneByUsername(ctx context.Context, username string) (res domain.User, err error) {
	query := `SELECT * FROM user WHERE username =? LIMIT 1`
	list, err := m.fetch(ctx, query, username)
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
func GenerateUUID() string {
	id := uuid.New()
	return id.String()
}
