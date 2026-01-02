package common

import (
	"github.com/gofiber/fiber/v2"
)

func GetSessionUserID(c *fiber.Ctx) string {
	userClaims, ok := c.Locals("user").(*JWTClaims)
	if !ok {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, GetRequestID(c))
		userID, _ := EncodedJwtToken(c, "user_id")
		return userID
	}

	return userClaims.UserID
}

func GetSession(c *fiber.Ctx) *JWTClaims {
	userClaims, ok := c.Locals("user").(*JWTClaims)
	if !ok {
		LogWarning("SESSION_INVALID_SIGNATURE", map[string]interface{}{"authorization": c.Get("Authorization")}, GetRequestID(c))
	}

	return userClaims
}
