package common

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

type JWTClaims struct {
	UserID string            `json:"user_id"`
	Role   string            `json:"role"`
	Remark map[string]string `json:"remark"`
	jwt.StandardClaims
}

func GenerateJWTToken(userID string, role string, remark map[string]string, expiresIn time.Duration) (string, error) {
	key := []byte(viper.GetString("jwt_secret_key"))

	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		Remark: remark,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresIn).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
