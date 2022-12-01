package domain

type CasbinRBAC struct {
	Sub string
	Obj string
	Act string
}

type CasbinRBACRepository interface {
	Store(c *CasbinRBAC) (bool, error)
}
