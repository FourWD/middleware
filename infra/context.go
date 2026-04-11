package infra

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	LocalRequestID    = "request_id"
	LocalAuthUser     = "auth_user_id"
	LocalAuthEmail    = "auth_email"
	LocalAuthRole     = "auth_role"
	LocalAuthNotiToken = "auth_noti_token"
)

func SetRequestID(c fiber.Ctx, requestID string) {
	c.Locals(LocalRequestID, requestID)
}

func GetRequestID(c fiber.Ctx) string {
	value, _ := c.Locals(LocalRequestID).(string)
	return value
}

func SetAuthContext(c fiber.Ctx, userID string, email, role string) {
	c.Locals(LocalAuthUser, userID)
	c.Locals(LocalAuthEmail, email)
	c.Locals(LocalAuthRole, role)
}

func SetAuthNotiToken(c fiber.Ctx, notiToken string) {
	c.Locals(LocalAuthNotiToken, notiToken)
}

func GetAuthNotiToken(c fiber.Ctx) string {
	value, _ := c.Locals(LocalAuthNotiToken).(string)
	return value
}

func GetAuthUserID(c fiber.Ctx) (string, bool) {
	value, ok := c.Locals(LocalAuthUser).(string)
	return value, ok
}

func GetAuthEmail(c fiber.Ctx) string {
	value, _ := c.Locals(LocalAuthEmail).(string)
	return value
}

func GetAuthRole(c fiber.Ctx) string {
	value, _ := c.Locals(LocalAuthRole).(string)
	return value
}

func AcceptLanguage(c fiber.Ctx) string {
	lang := c.Get("accept-language", "TH")
	return strings.ToUpper(lang)
}
