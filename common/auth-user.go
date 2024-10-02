package common

import (
	"context"
	"errors"
	"strings"

	"github.com/FourWD/middleware/orm"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

// var (
// 	mu        sync.RWMutex
// 	blacklist []string
// )

type UserAuthorization struct {
	IsSuccess bool
	Code      string
	Message   string
}

func CheckUserAuthorization(c *fiber.Ctx, db *gorm.DB, excludePath ...[]string) UserAuthorization {
	path := getLastPathComponent(c.Path())
	defaultExcludePath := []string{"login", "logout", "register", "wake-up", "warmup"}
	// Print("CheckUserAuthorization", path)

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
		// PrintError("LogUserLogin invalid request", token)
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "invalid request"}
	}

	var logUserLogin orm.LogUserLogin
	if err := db.Where("token = ?", token).Order("created_at DESC").First(&logUserLogin).Error; err != nil {
		// PrintError("LogUserLogin not found", token)
		return UserAuthorization{IsSuccess: false, Code: "401", Message: "log_login not found"}
	}

	if logUserLogin.ID == "" {
		// PrintError("LogUserLogin unauthorized", token)
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

func Login(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return errors.New("no token")
	}

	userID := GetSessionUserID(c)
	if err := deletePreviousLoginToken(userID); err != nil {
		return err
	}

	collection := DatabaseMongoMiddleware.Database.Collection("login_tokens")
	data := bson.M{
		"token":     token,
		"user_id":   userID,
		"issuedAt":  GetSession(c).IssuedAt,
		"expiresAt": GetSession(c).ExpiresAt,
	}

	_, err := collection.InsertOne(context.TODO(), data)
	return err
}

func Logout(c *fiber.Ctx) error {
	userID := GetSessionUserID(c)
	return deletePreviousLoginToken(userID)
}

func deletePreviousLoginToken(userID string) error {
	collection := DatabaseMongoMiddleware.Database.Collection("login_tokens")
	filter := bson.M{"user_id": userID}
	_, err := collection.DeleteMany(context.TODO(), filter)
	return err
}

// func addJwtBlacklist(token string) error {
// 	log.Println("addJwtBlacklist:", token)
// 	mu.Lock()
// 	defer mu.Unlock()

// 	// Check if the blacklist has reached its max size
// 	maxBlacklistSize := 150
// 	if len(blacklist) >= maxBlacklistSize {
// 		// Remove the oldest token (first in the slice)
// 		blacklist = blacklist[1:]
// 	}

// 	// Add the new token to the end of the slice
// 	blacklist = append(blacklist, token)
// 	return nil
// }
