package common

import (
	"math/rand"
	"time"

	"github.com/FourWD/middleware/kit"
)

func parseWithTimezone(strDateTime string, format string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		LogError("STRING_PARSE_ERROR", map[string]interface{}{"error": err.Error(), "input": strDateTime}, "")
		return kit.NilDate(), err
	}

	parsedTime, err := time.ParseInLocation(format, strDateTime, location)
	if err != nil {
		LogError("STRING_PARSE_ERROR", map[string]interface{}{"error": err.Error(), "input": strDateTime}, "")
		return kit.NilDate(), err
	}

	return parsedTime, nil
}

func StringToDateTime(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_MINUTE)
}

func StringToDate(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_DAY)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
