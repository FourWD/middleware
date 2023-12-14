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
	//fmt.Printf("Random string: %s\n", randomString)
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
	return t.Format("2006-01-02 15:04:05")
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

func StringToDateTime(strDateTime string) time.Time {
	layout := "2006-12-01 15:04"
	loc, _ := time.LoadLocation("Asia/Bangkok")
	dt, _ := time.ParseInLocation(layout, strDateTime, loc)
	return dt
}

func StringToDate(strDateTime string) time.Time {
	layout := "2006-12-01"
	loc, _ := time.LoadLocation("Asia/Bangkok")
	dt, _ := time.ParseInLocation(layout, strDateTime, loc)
	return dt
}
