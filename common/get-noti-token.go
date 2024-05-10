package common

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func GetNotiToken(c *fiber.Ctx) (string, error) {
	notiToken, _ := EncodedJwtToken(c, "noti_token") // for old login

	// if notiToken == "" {
	// 	session := GetSession(c)
	// 	if value, ok := session.Remark["noti_token"]; ok {
	// 		return value, nil
	// 	}
	// }

	if notiToken == "" {
		userClaims := c.Locals("user").(*JWTClaims)
		if value, ok := userClaims.Remark["noti_token"]; ok {
			return value, nil
		}
	}

	return "", errors.New("notiToken is nil")
}
