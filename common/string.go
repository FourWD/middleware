package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

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

func MD5(text string) string {
	hashText := md5.New()
	hashText.Write([]byte(text))
	//Print(" md5 pass : ", hex.EncodeToString(hashPassword.Sum(nil)))
	return hex.EncodeToString(hashText.Sum(nil))
}

func IsUUID(input string) bool {
	// Regular expression to match UUID format
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(input)
}
