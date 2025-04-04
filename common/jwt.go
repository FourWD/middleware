package common

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

// func AuthenticationMiddleware(c *fiber.Ctx) error {
// 	if isPublicPathV2(c) {
// 		// log.Println("public path")
// 		return c.Next()
// 	}
// 	return checkAuth(c)
// }

func AuthenticationMiddleware(c *fiber.Ctx) error {
	// FiberLog(c)
	if isPublicPath(c) {
		// log.Println("public path")
		return c.Next()
	}
	return checkAuth(c)
}

// func isPublicPath(c *fiber.Ctx) bool {
// 	publicPaths := viper.GetStringSlice("public_path")
// 	// log.Println("full_path:", c.Path())
// 	path := getLastPathComponent(c.Path())
// 	// log.Println("path:", path)
// 	// log.Println("publicPaths:", publicPaths)
// 	return StringExistsInList(path, publicPaths)
// }

func isPublicPath(c *fiber.Ctx) bool {
	publicPaths := viper.GetStringSlice("public_path")
	// log.Println("full_path:", c.Path())
	return StringExistsInList(c.Path(), publicPaths)
}

func checkAuth(c *fiber.Ctx) error {
	// log.Println("checkAuth (private path)")
	// Extract token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).SendString("No token provided")
	}

	// Check Blacklist
	if !IsJwtValid(authHeader) {
		return c.Status(http.StatusUnauthorized).SendString("token blacklist")
	}

	tokenString := authHeader[len("Bearer "):]

	// Parse the token
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwt_secret_key")), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return c.Status(http.StatusUnauthorized).SendString("Invalid token signature")
		}
		return c.Status(http.StatusBadRequest).SendString("Bad request")
	}

	if !token.Valid {
		return c.Status(http.StatusUnauthorized).SendString("Invalid token")
	}

	// Token is valid, do something with the claims
	c.Locals("user", claims)
	return c.Next()
}

// func IsJwtValid(token string) bool {
// 	var bl orm.JwtBlacklist
// 	result := Database.Model(orm.JwtBlacklist{}).Where("md5 = ?", MD5(token)).First(&bl)
// 	return result.RowsAffected == 0
// }

func IsJwtValid(token string) bool {
	collection := DatabaseMongoMiddleware.Database.Collection("blacklist_tokens")
	filter := bson.M{"token": token}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}

	if count > 0 {
		return false
	}

	return true
}

// func getLastPathComponent(path string) string {
// 	components := strings.Split(path, "/")
// 	lastComponent := components[len(components)-1]
// 	if lastComponent == "favicon.ico" {
// 		return ""
// 	}
// 	return lastComponent
// }

// func refreshTokenHandler(c *fiber.Ctx) error {
// 	// Extract refresh token from the request
// 	refreshToken := c.FormValue("refresh_token")
// 	if refreshToken == "" {
// 		return c.Status(http.StatusBadRequest).SendString("Refresh token not provided")
// 	}

// 	// Parse the refresh token
// 	claims := &JWTClaims{}
// 	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
// 		return refreshSecret, nil
// 	})

// 	if err != nil {
// 		if err == jwt.ErrSignatureInvalid {
// 			return c.Status(http.StatusUnauthorized).SendString("Invalid refresh token signature")
// 		}
// 		return c.Status(http.StatusBadRequest).SendString("Bad request")
// 	}

// 	if !token.Valid {
// 		return c.Status(http.StatusUnauthorized).SendString("Invalid refresh token")
// 	}

// 	// Refresh token is valid, generate a new access token
// 	newAccessToken, err := GenerateJWTToken(claims.UserID, secretKey, time.Hour*24)
// 	if err != nil {
// 		return err
// 	}

// 	return c.JSON(fiber.Map{"access_token": newAccessToken})
// }

// func FiberLogin(app *fiber.App) {
// 	app.Post("/login", func(c *fiber.Ctx) error {
// 		fmt.Println("mid", viper.GetString("jwt_secret_key"))
// 		fmt.Println("mid", viper.GetString("jwt_refresh_secret_key"))

// 		var user orm.User
// 		if err := c.BodyParser(&user); err != nil {
// 			return err
// 		}

// 		if authenticatedUser, ok := authenticate(user.Username, user.Password); ok {
// 			accessToken, err := GenerateJWTToken(authenticatedUser.ID, []byte(viper.GetString("jwt_secret_key")), time.Hour*24)
// 			if err != nil {
// 				return err
// 			}

// 			refreshToken, err := GenerateJWTToken(authenticatedUser.ID, []byte(viper.GetString("jwt_refresh_secret_key")), time.Hour*24*30) // 30 days
// 			if err != nil {
// 				return err
// 			}

// 			return c.JSON(fiber.Map{"token": accessToken, "refresh_token": refreshToken})
// 		}

// 		return c.Status(http.StatusUnauthorized).SendString("Invalid credentials")
// 	})
// }

// func FiberRefreshToken(c *fiber.Ctx) error {
// 	refreshToken := c.FormValue("refresh_token")
// 	if refreshToken == "" {
// 		return c.Status(http.StatusBadRequest).SendString("Refresh token not provided")
// 	}

// 	claims := &JWTClaims{}
// 	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
// 		return viper.GetString("jwt_refresh_secret_key"), nil
// 	})

// 	if err != nil {
// 		if err == jwt.ErrSignatureInvalid {
// 			return c.Status(http.StatusUnauthorized).SendString("Invalid refresh token signature")
// 		}
// 		return c.Status(http.StatusBadRequest).SendString("Bad request")
// 	}

// 	if !token.Valid {
// 		return c.Status(http.StatusUnauthorized).SendString("Invalid refresh token")
// 	}

// 	// Refresh token is valid, generate a new access token
// 	newAccessToken, err := GenerateJWTToken(claims.UserID, []byte(viper.GetString("jwt_secret_key")), time.Hour*24)
// 	if err != nil {
// 		return err
// 	}

// 	return c.JSON(fiber.Map{"token": newAccessToken})
// }

// func FiberTestProtection(app *fiber.App) {
// 	app.Get("/test-protection", func(c *fiber.Ctx) error {
// 		userClaims := c.Locals("user").(*JWTClaims)
// 		return c.JSON(fiber.Map{"user_id": userClaims.UserID})
// 	})
// }

// func authenticate(username, password string) (orm.User, bool) {
// 	var user orm.User
// 	result := Database.Where("username = ? AND password = ?", username, HashPassword(password)).First(&user)
// 	if result.Error != nil {
// 		return orm.User{}, false
// 	}

// 	return user, true
// }

// func authenticate(username, password string) (orm.User, bool) {
// 	// Replace this with your actual authentication logic
// 	// For simplicity, return a predefined user if credentials are valid
// 	if username == "user" && password == "password" {
// 		user := new(orm.User)
// 		user.ID = "1234"
// 		user.Username = "Admin"
// 		return *user, true
// 	}
// 	return orm.User{}, false
// }

// func main() {
// 	app := fiber.New()

// 	// Use authentication middleware for all routes except excluded ones
// 	app.Use(authenticationMiddleware)

// 	// Example login endpoint
// 	app.Post("/login", func(c *fiber.Ctx) error {
// 		var user User
// 		if err := c.BodyParser(&user); err != nil {
// 			return err
// 		}

// 		if authenticatedUser, ok := authenticate(user.Username, user.Password); ok {
// 			accessToken, err := generateToken(authenticatedUser.ID, authenticatedUser.Username, secretKey, time.Hour*24)
// 			if err != nil {
// 				return err
// 			}

// 			refreshToken, err := generateToken(authenticatedUser.ID, authenticatedUser.Username, refreshSecret, time.Hour*24*30) // 30 days
// 			if err != nil {
// 				return err
// 			}

// 			return c.JSON(fiber.Map{"access_token": accessToken, "refresh_token": refreshToken})
// 		}

// 		return c.Status(http.StatusUnauthorized).SendString("Invalid credentials")
// 	})

// 	// Example protected endpoint
// 	app.Get("/protected", func(c *fiber.Ctx) error {
// 		// Access the user from the authentication middleware
// 		userClaims := c.Locals("user").(*JWTClaims)
// 		return c.JSON(fiber.Map{"user_id": userClaims.UserID, "username": userClaims.Username})
// 	})

// 	// Refresh token endpoint
// 	app.Post("/refresh", refreshTokenHandler)

// 	app.Listen(":3000")
// }

func DecodeJWT(ResponseJwt string, tokenString string) (map[string]interface{}, error) {
	customClaims := make(map[string]interface{})

	token, err := jwt.Parse(ResponseJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenString), nil
	})

	if err != nil {
		return customClaims, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		for key, value := range claims {
			customClaims[key] = value
		}
		return customClaims, nil
	}

	return customClaims, err
}
