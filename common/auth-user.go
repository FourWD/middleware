package common

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserAuthorizationResult struct {
	IsSuccess bool
	Code      string
	Message   string
}

func CheckUserAuthorization(c *fiber.Ctx, db *gorm.DB) UserAuthorizationResult {
	bearerToken := c.Get("Authorization")
	token := strings.Replace(bearerToken, "Bearer ", "", 1)

	if token == "" {
		return UserAuthorizationResult{IsSuccess: false, Code: "401", Message: "INVALID REQUEST"}
	}

	sql := fmt.Sprintf(`SELECT * FROM users INNER JOIN log_logins ON users.id = log_logins.user_id 
	WHERE token = "%s" ORDER BY log_logins.created_at DESC LIMIT 1`, token)
	type UserToken struct {
		ID string `json:"id"`
	}
	userToken := new(UserToken)
	db.Raw(sql).Scan(&userToken)

	if userToken.ID == "" {
		return UserAuthorizationResult{IsSuccess: false, Code: "401", Message: "UNAUTHORIZED"}
	}

	return UserAuthorizationResult{IsSuccess: true, Code: "200", Message: "OK"}
}
