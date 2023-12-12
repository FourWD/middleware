package common

import (
	"fmt"
	"strings"

	"github.com/FourWD/middleware/orm"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserAuthorization struct {
	IsSuccess bool
	Code      string
	Message   string
}

func CheckUserAuthorization(c *fiber.Ctx, db *gorm.DB, excludePath ...[]string) UserAuthorization {
	path := getLastPathComponent(c.Path())
	if path == "login" || path == "logout" || path == "register" || StringExistsInList(path, excludePath[0]) {
		return UserAuthorization{IsSuccess: true, Code: "200", Message: "ok"}
	}

	bearerToken := c.Get("Authorization")
	token := strings.Replace(bearerToken, "Bearer ", "", 1)
	if token == "" {
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "invalid request"}
	}

	sql := fmt.Sprintf(`SELECT * FROM users INNER JOIN log_user_logins ON users.id = log_user_logins.user_id 
	WHERE token = "%s" ORDER BY log_user_logins.created_at DESC LIMIT 1`, token)
	logLogin := new(orm.LogUserLogin)
	db.Raw(sql).Scan(&logLogin)

	if logLogin.ID == "" {
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "unauthorized"}
	}

	return UserAuthorization{IsSuccess: true, Code: "200", Message: "ok"}
}

func getLastPathComponent(path string) string {
	components := strings.Split(path, "/")
	lastComponent := components[len(components)-1]
	return lastComponent
}
