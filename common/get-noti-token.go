package common

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func GetNotiToken(c *fiber.Ctx) (string, error) {
	notiToken, _ := EncodedJwtToken(c, "noti_token")

	if notiToken == "" {
		session := GetSession(c)
		if session == nil {
			return "", nil
		}

		if value, ok := session.Remark["noti_token"]; ok {
			Log("GET_NOTI_TOKEN", map[string]interface{}{"noti_token": value}, "")
			return value, nil
		}

		return "", errors.New("notiToken is nil")
	}

	return notiToken, nil
}
