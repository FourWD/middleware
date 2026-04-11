package common

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/FourWD/middleware/kit"
)

func DateString() string {
	currentTime := time.Now()
	dateString := fmt.Sprintf("%d", currentTime.UnixNano())
	randomDigits := generateRandomDigits(10)
	return dateString + randomDigits
}

func generateRandomDigits(count int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := ""
	for i := 0; i < count; i++ {
		result += fmt.Sprintf("%d", r.Intn(10))
	}
	return result
}

func DateToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DATE_FORMAT_SECOND)
}

func StringExistsInList(target string, strList []string) bool {
	for _, str := range strList {
		if str == target {
			return true
		}
	}
	return false
}

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
