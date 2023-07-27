package token

import (
	"errors"
	"fmt"
	"log"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// ParseJWT parses the given JWT token and returns the token claims.
func ParseJWT(tokenString string) (*jwt.Token, error) {
	signingKey := []byte(viper.GetString(`secret.jwt`))
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("unexpected jwt signing method: %v", t.Header["alg"])
		}
		return signingKey, nil
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	log.Println("ERRORORR:", err)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
func ParseRefreshToken(tokenString string) (*jwt.Token, error) {
	signingKeyRefreshToken := []byte(viper.GetString(`secret.refresh_jwt`))
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("unexpected jwt signing method: %v", t.Header["alg"])
		}
		return signingKeyRefreshToken, nil
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return token, nil
}
