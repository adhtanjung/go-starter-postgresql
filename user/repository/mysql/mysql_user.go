package mysql

import (
	"context"

	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/adhtanjung/go-boilerplate/pkg/helpers"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mysqlUserRepository struct {
	Conn   *gorm.DB
	Casbin *casbin.Enforcer
}
type mysqlUserFilepathRepository struct {
	Conn *gorm.DB
}

func NewMysqlUserRepository(Conn *gorm.DB, Casbin *casbin.Enforcer) domain.UserRepository {
	return &mysqlUserRepository{Conn, Casbin}
}
func NewMysqlUserFilepathRepository(Conn *gorm.DB) domain.UserFilepathRepository {
	return &mysqlUserFilepathRepository{Conn}
}

func (m *mysqlUserFilepathRepository) Store(ctx context.Context, u *domain.UserFilepath) (err error) {
	if result := m.Conn.Create(&u); result.Error != nil {
		return result.Error
	}
	return
}

func (m *mysqlUserRepository) Store(ctx context.Context, u *domain.User) (err error) {
	conn := m.Conn.WithContext(ctx)
	if result := conn.Create(&u); result.Error != nil {
		return result.Error
	}
	return
}

func (m *mysqlUserRepository) Update(ctx context.Context, u *domain.User) (err error) {
	if result := m.Conn.Model(&u).Updates(domain.User{Username: u.Username, Email: u.Email, Name: u.Name, Password: u.Password, ProfilePic: u.ProfilePic}); result.Error != nil {
		return result.Error
	}
	return

}

func (m *mysqlUserRepository) GetByID(ctx context.Context, id uuid.UUID) (res domain.User, err error) {
	var user domain.User
	m.Conn.First(&user, "id = ?", id)
	res = user
	return
}

func (m *mysqlUserRepository) GetOneByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (res domain.User, err error) {
	conn := m.Conn.WithContext(ctx)
	var user []domain.User
	result := conn.Preload("UserRoles", func(tx *gorm.DB) *gorm.DB {
		return tx.Select("ID, user_id, role_id")
	}).Preload("UserRoles.Role", func(tx *gorm.DB) *gorm.DB {
		return tx.Select("ID, Name")
	}).Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).Select("id, password").Find(&user)
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	if result.RowsAffected <= 0 {
		return res, domain.ErrNotFound
	}
	res = user[0]
	return
}

func (m *mysqlUserRepository) GetOne(ctx context.Context, args domain.UserQueryArgs) (res domain.User, err error) {
	var user domain.User
	if result := m.Conn.Preload("UserRoles", func(tx *gorm.DB) *gorm.DB {
		return tx.Select(helpers.Ternary(args.SelectClause.UserRoles, "id, user_id, role_id"))
	}).Preload("UserRoles.Role", func(tx *gorm.DB) *gorm.DB {
		return tx.Select(helpers.Ternary(args.SelectClause.Role, "id, name"))
	}).Where(args.WhereClause.User.Clause, args.WhereClause.User.Args).Select(helpers.Ternary(args.SelectClause.User, "id, username")).Find(&user); result.Error != nil {
		return domain.User{}, result.Error
	}
	res = user
	return
}
