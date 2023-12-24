package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

func RandomString(length int) string {
	if length <= 0 {
		length = 16
	}
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Printf("Failed to generate random bytes: %v\n", err)
	}
	// Encode the random bytes as a base64 string.
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)
	return randomString
}

func DateString() string {
	currentTime := time.Now()
	return currentTime.Format("20060102150405")
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

func StringToFloat(value string, fieldName string) float64 {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}

func StringToDateTime(strDateTime string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("parsed error:", err)
		return NilDate(), err
	}

	parsedTime, err := time.ParseInLocation(DATE_FORMAT_MINUTE, strDateTime, location)
	if err != nil {
		fmt.Println("parsed error:", err)
		return NilDate(), err
	}

	return parsedTime, nil
}

func StringToDate(strDateTime string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("parsed error:", err)
		return NilDate(), err
	}

	parsedTime, err := time.ParseInLocation(DATE_FORMAT_DAY, strDateTime, location)
	if err != nil {
		fmt.Println("parsed error:", err)
		return NilDate(), err
	}

	return parsedTime, nil
}
