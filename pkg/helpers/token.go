package helpers

import (
	"time"

	"github.com/adhtanjung/go-starter/domain"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type JwtCustomClaims struct {
	UserID uuid.UUID         `json:"id"`
	Roles  []domain.UserRole `json:"roles"`
	jwt.RegisteredClaims
}

type ShouldClaims struct {
	ExpiresAt int64
	Secret    string
}

func GenerateToken(userID uuid.UUID, userRoles []domain.UserRole, claimsParam ShouldClaims) (generatedToken string, err error) {
	hours := 24
	secret := viper.GetString(`secret.jwt`)

	dynamicClaims := ShouldClaims{
		int64(time.Hour * time.Duration(Ternary(claimsParam.ExpiresAt, int64(hours)))),
		Ternary(claimsParam.Secret, secret),
	}
	// convert the struct to a map
	m := map[string]interface{}{
		"ID":    userID,
		"Roles": userRoles,
	}
	user := domain.User{
		Base:      domain.Base{ID: m["ID"].(uuid.UUID)},
		UserRoles: m["Roles"].([]domain.UserRole),
	}
	claims := &JwtCustomClaims{
		user.ID,
		user.UserRoles,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(dynamicClaims.ExpiresAt))),
		},
	}
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	generatedToken, err = token.SignedString([]byte(dynamicClaims.Secret))
	if err != nil {
		return "", err
	}
	return
}

// func Generate2Tokens() (accessToken string, refreshToken string, err error) {
// 	hours := 24
// 	secret := viper.GetString(`secret.jwt`)

// 	dynamicClaims := ShouldClaims{
// 		int64(time.Hour * time.Duration(Ternary(claimsParam.ExpiresAt, int64(hours)))),
// 		Ternary(claimsParam.Secret, secret),
// 	}
// 	// convert the struct to a map
// 	m := map[string]interface{}{
// 		"ID":    userID,
// 		"Roles": userRoles,
// 	}
// 	user := domain.User{
// 		Base:      domain.Base{ID: m["ID"].(uuid.UUID)},
// 		UserRoles: m["Roles"].([]domain.UserRole),
// 	}
// 	claims := &JwtCustomClaims{
// 		user.ID,
// 		user.UserRoles,
// 		jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(dynamicClaims.ExpiresAt))),
// 		},
// 	}
// 	// Create token with claims
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	// Generate encoded token and send it as response.
// 	generatedToken, err := token.SignedString([]byte(dynamicClaims.Secret))
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return

// }
