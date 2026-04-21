package common

import (
	"time"

	"github.com/FourWD/middleware/infra"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims is an alias for infra.JWTClaims, kept for backwards compatibility.
type JWTClaims = infra.JWTClaims

func GenerateJWTToken(userID string, role string, remark map[string]string, expiresIn time.Duration) (string, error) {
	key := []byte(infra.GetEnv("JWT_SECRET", ""))

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
