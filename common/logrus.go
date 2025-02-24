package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	logrus "github.com/sirupsen/logrus"
)

func FiberLogrus(c *fiber.Ctx) error {
	requestID := uuid.New().String()
	c.Locals("request_id", requestID)

	startTime := time.Now()
	c.Locals("start_time", startTime)
	// latency := time.Since(created)

	authHeader := c.Get("Authorization")
	var jwtToken string
	var jwtClaims map[string]interface{}

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

	fields := logrus.Fields{
		"request_id":  requestID,
		"method":      c.Method(),
		"path":        c.Path(),
		"status_code": c.Response().StatusCode(),
		"body":        jsonBody,
		"jwt_decode":  jwtClaims,
	} // "ip":         c.IP(), 		// "latency":     latency.String(),
	AppLog.WithFields(fields).Info("REQUEST_STARTED")

	return c.Next()
}

func LogrusInfo(label string, fields logrus.Fields) {
	fields["created"] = time.Now().Format(DATE_FORMAT_NANO)
	fields["status"] = 1
	fields["message"] = "success"
	AppLog.WithFields(fields).Info(label)
}

func LogrusError(label string, fields logrus.Fields, err error) {
	fields["created"] = time.Now().Format(DATE_FORMAT_NANO)
	fields["status"] = 0
	fields["message"] = err.Error()
	AppLog.WithFields(fields).Error(label)
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

	fields := logrus.Fields{
		"request_id": requestID,
		"duration":   duration.Milliseconds(),
	}

	AppLog.WithFields(fields).Info("REQUEST_COMPLETE")
}
