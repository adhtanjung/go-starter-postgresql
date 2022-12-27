package middlewares

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func TokenToContext(secret string) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqToken := c.Request().Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			reqToken = splitToken[1]
			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(reqToken, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil {
				logrus.Error("invalid token")
				return echo.ErrForbidden
			}

			userID := claims["id"]
			c.Set("user_id", userID)

			return next(c)
		}
	}

}
