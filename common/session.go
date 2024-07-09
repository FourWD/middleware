package common

import "github.com/gofiber/fiber/v2"

func GetSessionUserID(c *fiber.Ctx) string {
	userClaims := c.Locals("user").(*JWTClaims)
	return userClaims.UserID //
}

func GetSession(c *fiber.Ctx) *JWTClaims {
	userClaims := c.Locals("user").(*JWTClaims)
	return userClaims
}
