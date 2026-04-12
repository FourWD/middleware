package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func FiberLog(c fiber.Ctx) error {
	requestID := uuid.New().String()
	c.Locals("request_id", requestID)

	startTime := time.Now()
	c.Locals("start_time", startTime)

	authHeader := c.Get("Authorization")
	var jwtClaims map[string]interface{} = make(map[string]interface{})

	if strings.HasPrefix(authHeader, "Bearer ") {
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
		claimsJson, err := decodeToJson(jwtToken)
		if err == nil {
			jwtClaims = claimsJson
		}
	}

	body := c.Body()
	var jsonBody map[string]interface{}

	if err := json.Unmarshal(body, &jsonBody); err != nil {
		jsonBody = map[string]interface{}{"raw_body": string(body)}
	}

	AppLog.Info(infra.M("REQUEST_STARTED"),
		infra.WithField("request_id", requestID),
		infra.WithField("method", c.Method()),
		infra.WithField("path", c.Path()),
		infra.WithField("status", c.Response().StatusCode()),
		infra.WithDataField("body", jsonBody),
		infra.WithDataField("jwt_decode", jwtClaims),
	)
	return c.Next()
}

func GetRequestID(c fiber.Ctx) string {
	requestID := fiber.Locals[string](c, "request_id")
	return requestID
}

func Log(label string, logData map[string]interface{}, requestID string) {
	AppLog.Info(infra.M(label),
		infra.WithField("request_id", requestID),
		infra.WithDataFields(logData),
	)
}

func LogWarning(label string, logData map[string]interface{}, requestID string) {
	AppLog.Warn(infra.M(label),
		infra.WithField("request_id", requestID),
		infra.WithDataFields(logData),
	)
}

func LogError(label string, logData map[string]interface{}, requestID string) {
	AppLog.Error(nil, infra.M(label),
		infra.WithField("request_id", requestID),
		infra.WithDataFields(logData),
	)
}

func decodeToJson(jwtToken string) (map[string]interface{}, error) {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		claimsJson := make(map[string]interface{})
		for key, value := range claims {
			claimsJson[key] = value
		}
		return claimsJson, nil
	}
	return nil, nil
}

func responseLog(c fiber.Ctx) {
	requestID := GetRequestID(c)
	startTime := fiber.Locals[time.Time](c, "start_time")
	duration := time.Since(startTime)

	AppLog.Info(infra.M("REQUEST_COMPLETE"),
		infra.WithField("request_id", requestID),
		infra.WithField("duration_ms", duration.Milliseconds()),
	)
}
