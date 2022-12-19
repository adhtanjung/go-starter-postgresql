package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

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
	_ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/adhtanjung/go-boilerplate/domain"
	middlewares "github.com/adhtanjung/go-boilerplate/pkg/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.elastic.co/apm/module/apmechov4/v2"

	// _articleHttpDelivery "github.com/adhtanjung/go-boilerplate/article/delivery/http"
	_articleHttpDeliveryMiddleware "github.com/adhtanjung/go-boilerplate/article/delivery/http/middleware"
	// _articleRepo "github.com/adhtanjung/go-boilerplate/article/repository/mysql"
	// _articleUcase "github.com/adhtanjung/go-boilerplate/article/usecase"
	_authHttpDelivery "github.com/adhtanjung/go-boilerplate/auth/delivery/http"
	_authUcase "github.com/adhtanjung/go-boilerplate/auth/usecase"

	_casbinRepo "github.com/adhtanjung/go-boilerplate/casbin/repository/mysql"
	_roleHttpDelivery "github.com/adhtanjung/go-boilerplate/role/delivery/http"
	_roleRepo "github.com/adhtanjung/go-boilerplate/role/repository/mysql"
	_roleUcase "github.com/adhtanjung/go-boilerplate/role/usecase"
	_userHttpDelivery "github.com/adhtanjung/go-boilerplate/user/delivery/http"
	_userRepo "github.com/adhtanjung/go-boilerplate/user/repository/mysql"
	_userUcase "github.com/adhtanjung/go-boilerplate/user/usecase"
	_userRoleRepo "github.com/adhtanjung/go-boilerplate/user_role/repository/mysql"
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
		}

		for _, role := range roles {
			result, err := e.enforcer.Enforce(role, path, method)
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

func hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	//
	defer ws.Close()

	for {
		// Write
		err := ws.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
		if err != nil {
			c.Logger().Error(err)
		}

		// Read
		_, msg, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error(err)
		}
		fmt.Printf("%s\n", msg)
	}
}

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
	// key := "Secret-session-key" // Replace with your SESSION_SECRET or similar
	// maxAge := 86400 * 30        // 30 days
	// isProd := false             // Set to true when serving over https

	// store := sessions.NewCookieStore([]byte(key))
	// store.MaxAge(maxAge)
	// store.Options.Path = "/"
	// store.Options.HttpOnly = true // HttpOnly should always be enabled
	// store.Options.Secure = isProd

	// gothic.Store = store
	goth.UseProviders(
		google.New(viper.GetString("google.client_id"), viper.GetString("google.client_secret"), "http://127.0.0.1:9090/auth/callback?provider=google"),
	)
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}
	if err := dbConn.AutoMigrate(&domain.User{}, &domain.UserFilepath{}, &domain.UserRole{}); err != nil {
		log.Println("migration error", err)
	}
	casbinDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/boilerplate", dbUser, dbPass, dbHost, dbPort)
	a, _ := gormadapter.NewAdapter("mysql", casbinDsn, true) // Your driver and data source.

	en, _ := casbin.NewEnforcer("auth_model.conf", a)
	enforcer := Enforcer{enforcer: en}

	// Load the policy from DB.
	en.LoadPolicy()

	signingKey := []byte(viper.GetString(`secret.jwt`))

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
		return c.HTML(http.StatusOK, `
			<h1>Welcome to Softworx API!</h1>
		`)
	})

	e.Static("/", "public")
	middL := _articleHttpDeliveryMiddleware.InitMiddleware()
	// e.Use(middleware.Logger())
	e.GET("/ws", hello)
	e.Use(middleware.Recover())
	e.Use(middlewares.MiddlewareLogging)
	e.Use(apmechov4.Middleware())
	// e.GET("/auth/google/callback", func(c echo.Context) error{
	// 	return c.
	// })
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
		// providerName := c.QueryParam("provider")
		// provider, err := goth.GetProvider(providerName)
		// if err != nil {
		// 	return err
		// }
		// sess, err := provider.BeginAuth("state")
		// if err != nil {
		// 	return err
		// }
		// url, err := sess.GetAuthURL()
		// if err != nil {
		// 	return err
		// }

		gothic.BeginAuthHandler(c.Response(), c.Request())
		// return c.Redirect(http.StatusFound, url)
		return nil
	})
	e.GET("/test-google", func(c echo.Context) error {
		t, _ := template.ParseFiles("./web/oauth_login.html")
		t.Execute(c.Response(), false)
		// return c.HTML(http.StatusOK, `<p><a href="/auth/google">Log in with google</a></p>`)
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

	apiGroup := e.Group("/api")
	v1 := apiGroup.Group("/v1")
	v1.Use(middleware.JWTWithConfig(config))
	v1.Use(middL.CORS)
	v1.Use(enforcer.Enforce)
	v1.Use(middlewares.TokenToContext)
	// authorRepo := _authorRepo.NewMysqlAuthorRepository(dbConn)
	// ar := _articleRepo.NewMysqlArticleRepository(dbConn)
	userRepo := _userRepo.NewMysqlUserRepository(dbConn, en)
	userFilepathRepo := _userRepo.NewMysqlUserFilepathRepository(dbConn)
	roleRepo := _roleRepo.NewMysqlRoleRepository(dbConn)
	userRoleRepo := _userRoleRepo.NewMysqlUserRoleRepository(dbConn)
	casbinRepo := _casbinRepo.NewMysqlCasbinRepository(en)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	us := _userUcase.NewUserUsecase(userRepo, roleRepo, userRoleRepo, casbinRepo, userFilepathRepo, timeoutContext)
	ru := _roleUcase.NewRoleUsecase(roleRepo, timeoutContext)
	auth := _authUcase.NewAuthUsecase(userRepo, userRoleRepo, roleRepo, timeoutContext)
	// au := _articleUcase.NewArticleUsecase(ar, authorRepo, timeoutContext)

	_authHttpDelivery.NewAuthHandler(e, auth)
	// _articleHttpDelivery.NewArticleHandler(e, au)
	_userHttpDelivery.NewUserHandler(v1, us)
	_roleHttpDelivery.NewRoleHandler(v1, ru)

	// s := http.Server{
	// 	Addr:      ":9090",
	// 	Handler:   e, // set Echo as handler
	// 	TLSConfig: tlsConfig,
	// }
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
