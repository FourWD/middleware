package common

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type JWTClaims struct {
	UserID string            `json:"user_id"`
	Role   string            `json:"role"`
	Remark map[string]string `json:"remark"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(userID string, role string, remark map[string]string, expiresIn time.Duration) (string, error) {
	key := []byte(viper.GetString("jwt_secret_key"))

	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		Remark: remark,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
