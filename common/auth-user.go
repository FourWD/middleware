package common

import (
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
	defaultExcludePath := []string{"login", "logout", "register", "wake-up", "warmup"}
	Print("CheckUserAuthorization", path)

	if StringExistsInList(path, defaultExcludePath) {
		return UserAuthorization{IsSuccess: true, Code: "200", Message: "ok"}
	}

	if len(excludePath) > 1 {
		if StringExistsInList(path, defaultExcludePath) || StringExistsInList(path, excludePath[0]) {
			return UserAuthorization{IsSuccess: true, Code: "200", Message: "ok"}
		}
	}

	bearerToken := c.Get("Authorization")
	token := strings.Replace(bearerToken, "Bearer ", "", 1)
	if token == "" {
		PrintError("LogUserLogin invalid request", token)
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "invalid request"}
	}

	var logUserLogin orm.LogUserLogin
	if err := db.Where("token = ?", token).Order("created_at DESC").First(&logUserLogin).Error; err != nil {
		PrintError("LogUserLogin not found", token)
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "log_login not found"}
	}

	if logUserLogin.ID == "" {
		PrintError("LogUserLogin unauthorized", token)
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "unauthorized"}
	}

	return UserAuthorization{IsSuccess: true, Code: "200", Message: "ok"}
}

func getLastPathComponent(path string) string {
	components := strings.Split(path, "/")
	lastComponent := components[len(components)-1]
	if lastComponent == "favicon.ico" {
		return ""
	}
	return lastComponent
}
