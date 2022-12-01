package mysql

import (
	"github.com/adhtanjung/go-boilerplate/domain"
	"github.com/casbin/casbin/v2"
)

type mysqlCasbinRepository struct {
	Casbin *casbin.Enforcer
}

func NewMysqlCasbinRepository(Casbin *casbin.Enforcer) domain.CasbinRBACRepository {
	return &mysqlCasbinRepository{Casbin}
}
func (m *mysqlCasbinRepository) Store(c *domain.CasbinRBAC) (b bool, err error) {
	b, err = m.Casbin.AddPolicy(c.Sub, c.Obj, c.Act)
	return

}
