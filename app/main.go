package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"runtime"
	_ "runtime/pprof"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/sirupsen/logrus"

	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	// _ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/adhtanjung/go-starter/domain"
	middlewares "github.com/adhtanjung/go-starter/pkg/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	// "go.elastic.co/apm/module/apmechov4/v2"

	// _articleHttpDelivery "github.com/adhtanjung/go-starter/article/delivery/http"
	// _articleRepo "github.com/adhtanjung/go-starter/article/repository/mysql"
	// _articleUcase "github.com/adhtanjung/go-starter/article/usecase"
	_authHttpDelivery "github.com/adhtanjung/go-starter/auth/delivery/http"
	_authUcase "github.com/adhtanjung/go-starter/auth/usecase"

	// _refreshTokenHttpDelivery "github.com/adhtanjung/go-starter/auth/delivery/http"
	// _casbinRepo "github.com/adhtanjung/go-starter/casbin/repository/mysql"
	_roleHttpDelivery "github.com/adhtanjung/go-starter/role/delivery/http"
	_roleRepo "github.com/adhtanjung/go-starter/role/repository/mysql"
	_roleUcase "github.com/adhtanjung/go-starter/role/usecase"
	_userHttpDelivery "github.com/adhtanjung/go-starter/user/delivery/http"
	_userRepo "github.com/adhtanjung/go-starter/user/repository/mysql"
	_userUcase "github.com/adhtanjung/go-starter/user/usecase"
	_userRoleRepo "github.com/adhtanjung/go-starter/user_role/repository/mysql"
) // TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	viper.SetConfigFile(`config.json`)
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

type Enforcer struct {
	enforcer *casbin.Enforcer
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
			result, err := e.enforcer.Enforce(role, path, method)
			log.Println("role:", role)
			log.Println("path:", path)
			log.Println("method:", method)
			log.Println("result:", result)
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

var (
	upgrader = websocket.Upgrader{}
)

type M map[string]interface{}
type Renderer struct {
	template *template.Template
	debug    bool
	location string
}

func NewRenderer(location string, debug bool) *Renderer {
	tpl := new(Renderer)
	tpl.location = location
	tpl.debug = debug

	tpl.ReloadTemplates()

	return tpl
}
func (t *Renderer) ReloadTemplates() {
	t.template = template.Must(template.ParseGlob(t.location))
}

func (t *Renderer) Render(
	w io.Writer,
	name string,
	data interface{},
	c echo.Context,
) error {
	if t.debug {
		t.ReloadTemplates()
	}

	return t.template.ExecuteTemplate(w, name, data)
}

func main() {
	runtime.GOMAXPROCS(4)

	gugel := google.New(viper.GetString("google.client_id"), viper.GetString("google.client_secret"), "http://localhost:9090/auth/callback?provider=google")
	gugel.SetPrompt("consent")

	goth.UseProviders(
		gugel,
	)
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	// connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	// connection := "postgres://adhitanjung:@localhost:5432/starter"
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", dbHost, dbUser, dbPass, dbName, dbPort)
	// dsn := "host=localhost user=adhitanjung password=asdqwe123 dbname=starter port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}
	if err := dbConn.AutoMigrate(&domain.User{}, &domain.UserFilepath{}, &domain.UserRole{}); err != nil {
		log.Println("migration error", err)
	}
	// casbinDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/boilerplate", dbUser, dbPass, dbHost, dbPort)
	casbinDsn := "host=localhost user=adhitanjung password=asdqwe123 dbname=casbin port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	a, _ := gormadapter.NewAdapter("postgres", casbinDsn, true) // Your driver and data source.

	en, _ := casbin.NewEnforcer("auth_model.conf", a)
	enforcer := Enforcer{enforcer: en}
	// en.AddPolicy("superadmin", "/*", "*")
	// en.AddPolicy("user", "/api/v1/users/*", "*")
	// en.AddPolicy("user", "/logout", "*")

	// Load the policy from DB.
	en.LoadPolicy()
	enforcc, _ := en.Enforce("user", "/api/v1/users", "POST")
	log.Println("WOOOOOW", enforcc)
	en.SavePolicy()
	// en.EnableAutoSave(true)

	signingKey := []byte(viper.GetString(`secret.jwt`))
	signingKeyRefreshToken := []byte(viper.GetString(`secret.refresh_jwt`))

	config := middleware.JWTConfig{
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return signingKey, nil
			}

			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := jwt.Parse(auth, keyFunc)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
	}
	configRefreshToken := middleware.JWTConfig{
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return signingKeyRefreshToken, nil
			}

			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := jwt.Parse(auth, keyFunc)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
	}
	e := echo.New()
	dir, err := os.Getwd()
	if err != nil {
		logrus.Error(err)
	}

	e.Renderer = NewRenderer(filepath.Join(dir, "/web/*.html"), true)

	// Add a healthcheck endpoint
	e.GET("/healthcheck", func(c echo.Context) error {
		// Return a 200 OK status code and a "healthy" message
		return c.String(http.StatusOK, "healthy")
	})
	e.GET("/", func(c echo.Context) error {
		c.SetCookie(&http.Cookie{
			Name:  "test_cookie",
			Value: "woy",
		})
		return c.HTML(http.StatusOK, `
			<h1>Welcome to Softworx API!</h1>
		`)
	})

	e.Static("/", "public")
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))
	e.Use(middleware.Recover())
	e.Use(middlewares.MiddlewareLogging)
	// e.Use(apmechov4.Middleware())
	e.GET("/auth/callback", func(c echo.Context) error {

		user, err := gothic.CompleteUserAuth(c.Response(), c.Request())

		if err != nil {
			fmt.Fprintln(c.Response(), err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		t, _ := template.ParseFiles("./web/oauth_success.html")
		t.Execute(c.Response(), user)
		return c.Render(http.StatusOK, "callback", t)

	})
	e.GET("/auth", func(c echo.Context) error {

		gothic.BeginAuthHandler(c.Response(), c.Request())
		return nil
	})
	e.GET("/test-google", func(c echo.Context) error {
		t, _ := template.ParseFiles("./web/oauth_login.html")
		t.Execute(c.Response(), false)
		return c.Render(http.StatusOK, "home", t)
	})

	// e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
	// 	LogURI:       true,
	// 	LogStatus:    true,
	// 	LogMethod:    true,
	// 	LogRemoteIP:  true,
	// 	LogUserAgent: true,
	// 	LogError:     true,
	// 	LogRoutePath: true,
	// 	LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
	// 		log.WithFields(logrus.Fields{
	// 			"method":     values.Method,
	// 			"URI":        values.URI,
	// 			"status":     values.Status,
	// 			"error":      values.Error,
	// 			"user_agent": values.UserAgent,
	// 			"remote_ip":  values.RemoteIP,
	// 			"route_path": values.RoutePath,
	// 		}).Info("request")
	// 		return nil
	// 	},
	// }))
	refreshToken := e.Group("/refresh-token")
	refreshToken.Use(middleware.JWTWithConfig(configRefreshToken))
	refreshToken.Use(middlewares.TokenToContext(viper.GetString("secret.refresh_jwt")))

	apiGroup := e.Group("/api")
	v1 := apiGroup.Group("/v1")
	v1.Use(middleware.JWTWithConfig(config))
	v1.Use(enforcer.Enforce)
	v1.Use(middlewares.TokenToContext(viper.GetString("secret.jwt")))
	// authorRepo := _authorRepo.NewMysqlAuthorRepository(dbConn)
	// ar := _articleRepo.NewMysqlArticleRepository(dbConn)
	userRepo := _userRepo.NewMysqlUserRepository(dbConn, en)
	userFilepathRepo := _userRepo.NewMysqlUserFilepathRepository(dbConn)
	roleRepo := _roleRepo.NewMysqlRoleRepository(dbConn)
	userRoleRepo := _userRoleRepo.NewMysqlUserRoleRepository(dbConn)
	// casbinRepo := _casbinRepo.NewMysqlCasbinRepository(en)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	// us := _userUcase.NewUserUsecase(userRepo, roleRepo, userRoleRepo, casbinRepo, userFilepathRepo, timeoutContext)
	us := _userUcase.NewUserUsecase(userRepo, roleRepo, userRoleRepo, userFilepathRepo, timeoutContext)
	ru := _roleUcase.NewRoleUsecase(roleRepo, timeoutContext)
	auth := _authUcase.NewAuthUsecase(userRepo, userRoleRepo, roleRepo, timeoutContext)
	// au := _articleUcase.NewArticleUsecase(ar, authorRepo, timeoutContext)

	_authHttpDelivery.NewRefreshTokenHandler(refreshToken, us)
	_authHttpDelivery.NewAuthHandler(e, auth)
	// _articleHttpDelivery.NewArticleHandler(e, au)
	_userHttpDelivery.NewUserHandler(v1, us)
	_roleHttpDelivery.NewRoleHandler(v1, ru)

	e.HTTPErrorHandler = middlewares.ErrorHandler
	lock := make(chan error)
	time.Sleep(1 * time.Millisecond)
	middlewares.MakeLogEntry(nil).Warning("application started without ssl/tls enabled")
	go func(lock chan error) { lock <- e.Start(viper.GetString("server.address")) }(lock)
	errN := <-lock
	if errN != nil {
		middlewares.MakeLogEntry(nil).Panic("failed to start application")
	}
	// if err := s.ListenAndServeTLS("server.crt", "server.key"); err != http.ErrServerClosed {
	// 	e.Logger.Fatal(err)
	// }
	// log.Fatal(e.Start(viper.GetString("server.address")))
}
