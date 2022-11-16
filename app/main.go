package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	_articleHttpDelivery "github.com/bxcodec/go-clean-arch/article/delivery/http"
	_articleHttpDeliveryMiddleware "github.com/bxcodec/go-clean-arch/article/delivery/http/middleware"
	_articleRepo "github.com/bxcodec/go-clean-arch/article/repository/mysql"
	_articleUcase "github.com/bxcodec/go-clean-arch/article/usecase"
	_authHttpDelivery "github.com/bxcodec/go-clean-arch/auth/delivery/http"
	_authUcase "github.com/bxcodec/go-clean-arch/auth/usecase"
	_authorRepo "github.com/bxcodec/go-clean-arch/author/repository/mysql"
	_userHttpDelivery "github.com/bxcodec/go-clean-arch/user/delivery/http"
	_userRepo "github.com/bxcodec/go-clean-arch/user/repository/mysql"
	_userUcase "github.com/bxcodec/go-clean-arch/user/usecase"
	middleware "github.com/labstack/echo/v4/middleware"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func main() {
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
	dbConn, err := sql.Open(`mysql`, dsn)

	if err != nil {
		log.Fatal(err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	signingKey := []byte("secret")

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
	middL := _articleHttpDeliveryMiddleware.InitMiddleware()
	apiGroup := e.Group("/api")
	apiGroup.Use(middleware.JWTWithConfig(config))
	apiGroup.Use(middL.CORS)
	authorRepo := _authorRepo.NewMysqlAuthorRepository(dbConn)
	ar := _articleRepo.NewMysqlArticleRepository(dbConn)

	userRepo := _userRepo.NewMysqlUserRepository(dbConn)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	us := _userUcase.NewUserUsecase(userRepo, timeoutContext)
	auth := _authUcase.NewAuthUsecase(userRepo, timeoutContext)
	au := _articleUcase.NewArticleUsecase(ar, authorRepo, timeoutContext)
	_authHttpDelivery.NewAuthHandler(e, auth)
	_articleHttpDelivery.NewArticleHandler(e, au)
	_userHttpDelivery.NewUserHandler(apiGroup, us)

	log.Fatal(e.Start(viper.GetString("server.address")))
}
