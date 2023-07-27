package server

import (
	"errors"
	"net/http"

	"fmt"
	"html/template"

	// "github.com/adhtanjung/go-starter/pkg/casbin"
	"github.com/adhtanjung/go-starter/pkg/middlewares"
	"github.com/adhtanjung/go-starter/pkg/token"

	// "github.com/golang-jwt/jwt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth/gothic"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	// _refreshTokenHttpDelivery "github.com/adhtanjung/go-starter/auth/delivery/http"
	// _casbinRepo "github.com/adhtanjung/go-starter/casbin/repository/mysql"
)

func (a *App) registerApiRoutes() {
	// enforcer := casbin.Enforcer{enforcer: en}
	config := middleware.JWTConfig{
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {

			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := token.ParseJWT(auth)
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
			token, err := token.ParseJWT(auth)
			if err != nil {
				return nil, err
			}
			return token, nil
		},
	}
	// Healthcheck endpoint
	a.e.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	// Other routes and handlers go here
	// For example:
	// a.e.GET("/", func(c echo.Context) error {
	// 	return c.HTML(http.StatusOK, "<h1>Welcome to the API</h1>")
	// })

	// Swagger documentation
	url := echoSwagger.URL("http://localhost:9090/swagger/doc.json") //The url pointing to API definition
	a.e.GET("/swagger/*", echoSwagger.EchoWrapHandler(url))

	// Handle other routes
	a.e.Static("/", "public")
	a.e.GET("/auth/callback", func(c echo.Context) error {

		user, err := gothic.CompleteUserAuth(c.Response(), c.Request())

		if err != nil {
			fmt.Fprintln(c.Response(), err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		t, _ := template.ParseFiles("./web/oauth_success.html")
		t.Execute(c.Response(), user)
		return c.Render(http.StatusOK, "callback", t)

	})
	a.e.GET("/auth", func(c echo.Context) error {

		gothic.BeginAuthHandler(c.Response(), c.Request())
		return nil
	})
	refreshToken := a.e.Group("/refresh-token")
	refreshToken.Use(middleware.JWTWithConfig(configRefreshToken))
	refreshToken.Use(middlewares.TokenToContext(viper.GetString("secret.refresh_jwt")))

	apiGroup := a.e.Group("/api")
	v1 := apiGroup.Group("/v1")
	v1.Use(middleware.JWTWithConfig(config))
	// v1.Use(enforcer)
	v1.Use(middlewares.TokenToContext(viper.GetString("secret.jwt")))

	a.e.HTTPErrorHandler = middlewares.ErrorHandler
}
