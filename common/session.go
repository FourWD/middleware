package common

import (
	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
)

func GetSessionUserID(c fiber.Ctx) string {
	userClaims := fiber.Locals[*JWTClaims](c, "user")
	if userClaims == nil {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, infra.GetRequestID(c))
		userID, _ := EncodedJwtToken(c, "user_id")
		return userID
	}

	return userClaims.UserID
}

func GetSession(c fiber.Ctx) *JWTClaims {
	userClaims := fiber.Locals[*JWTClaims](c, "user")
	if userClaims == nil {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, infra.GetRequestID(c))
	}

	return userClaims
}
