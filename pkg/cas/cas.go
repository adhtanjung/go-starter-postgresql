package cas

import (
	"log"
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type Enforcer struct {
	Enforcer *casbin.Enforcer
}

func (e *Enforcer) Enforce(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method
		path := c.Request().URL.Path
		reqToken := c.Request().Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		reqToken = splitToken[1]
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(reqToken, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString(`secret.jwt`)), nil
		})
		if err != nil {
			panic(err)
		}
		userId := claims["id"]
		if userId == nil {
			panic("user id not found")
		}
		rolesFromToken := claims["roles"]
		if rolesFromToken == nil {
			panic("no roles")
		}

		rolesSlc, okk := rolesFromToken.([]any)
		if !okk {
			panic("roles not a slice")
		}
		var roles []string
		for _, r := range rolesSlc {
			rMap, ok := r.(map[string]interface{})
			if !ok {
				panic("rmap not a map")
			}
			role := rMap["role"].(map[string]interface{})

			roles = append(roles, role["name"].(string))
			log.Println(role["name"])
		}

		for _, role := range roles {
			log.Println(role)
			log.Println(path)
			log.Println(method)
			result, err := e.Enforcer.Enforce(role, path, method)
			if err != nil {
				return echo.ErrInternalServerError
			}
			if result {
				return next(c)
			}
		}

		return echo.ErrForbidden
	}
}

func InitCasbin() (enforcer Enforcer, en *casbin.Enforcer) {
	casbinDsn := "host=localhost user=adhitanjung password=asdqwe123 dbname=casbin port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	a, _ := gormadapter.NewAdapter("postgres", casbinDsn, true) // Your driver and data source.

	en, _ = casbin.NewEnforcer("auth_model.conf", a)
	enforcer = Enforcer{Enforcer: en}
	en.AddPolicy("superadmin", "/*", "*")
	en.AddPolicy("user", "/api/v1/users/", "*")
	en.AddPolicy("user", "/logout", "*")

	// Load the policy from DB.
	en.LoadPolicy()
	en.SavePolicy()

	return enforcer, en

}
