package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func FiberLog(c *fiber.Ctx) error {
	requestID := uuid.New().String()
	c.Locals("request_id", requestID)

	startTime := time.Now()
	c.Locals("start_time", startTime)
	// latency := time.Since(created)

	authHeader := c.Get("Authorization")
	var jwtToken string
	var jwtClaims map[string]interface{} = make(map[string]interface{})

	if strings.HasPrefix(authHeader, "Bearer ") {
		jwtToken = strings.TrimPrefix(authHeader, "Bearer ")
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

	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Int("status_code", c.Response().StatusCode()),
		zap.Any("body", jsonBody),
		zap.Any("jwt_decode", jwtClaims),
	} // "ip":         c.IP(), 		// "latency":     latency.String(),
	AppLog.Info("REQUEST_STARTED", fields...)
	return c.Next()
}

func GetRequestID(c *fiber.Ctx) string {
	requestID, _ := c.Locals("request_id").(string)
	return requestID
}

func Log(label string, fields map[string]interface{}, requestID ...string) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	if len(requestID) > 0 {
		fields["request_id"] = requestID[0]
	}

	logFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		logFields = append(logFields, zap.Any(k, v))
	}

	AppLog.Info(label, logFields...)
}

func LogError(label string, fields map[string]interface{}, requestID ...string) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	if len(requestID) > 0 {
		fields["request_id"] = requestID[0]
	}

	logFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		logFields = append(logFields, zap.Any(k, v))
	}

	AppLog.Error(label, logFields...)
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

func responseLog(c *fiber.Ctx) {
	requestID, _ := c.Locals("request_id").(string)
	startTime, _ := c.Locals("start_time").(time.Time)
	duration := time.Since(startTime)

	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.Int64("duration", duration.Milliseconds()),
	}
	AppLog.Info("REQUEST_COMPLETE", fields...)
}
