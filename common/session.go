package common

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// func GetSessionUserID(c *fiber.Ctx) string {
// 	userClaims := c.Locals("user").(*JWTClaims)
// 	return userClaims.UserID //
// }

func GetSessionUserID(c *fiber.Ctx) string {
	userClaims, ok := c.Locals("user").(*JWTClaims)
	if !ok {
		log.Printf("Invalid signature [%s]", c.Get("Authorization"))
		userID, _ := EncodedJwtToken(c, "user_id")
		return userID
	}

	return userClaims.UserID
}

func GetSession(c *fiber.Ctx) *JWTClaims {
	userClaims, ok := c.Locals("user").(*JWTClaims)
	if !ok {
		log.Printf("Invalid signature [%s]", c.Get("Authorization"))
	}

	return userClaims
}
