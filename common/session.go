package common

import (
	"github.com/gofiber/fiber/v3"
)

func GetSessionUserID(c fiber.Ctx) string {
	userClaims := fiber.Locals[*JWTClaims](c, "user")
	if userClaims == nil {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, GetRequestID(c))
		userID, _ := EncodedJwtToken(c, "user_id")
		return userID
	}

	return userClaims.UserID
}

func GetSession(c fiber.Ctx) *JWTClaims {
	userClaims := fiber.Locals[*JWTClaims](c, "user")
	if userClaims == nil {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, GetRequestID(c))
	}

	return userClaims
}
