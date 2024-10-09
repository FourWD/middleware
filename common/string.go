package common

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
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
		result += fmt.Sprintf("%d", r.Intn(10)) // เลขสุ่มระหว่าง 0-9
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

func Hash(text string, salt string) string {

	hashText := sha256.New()
	hashText.Write([]byte(text + salt))
	//Print(" md5 pass : ", hex.EncodeToString(hashPassword.Sum(nil)))
	return hex.EncodeToString(hashText.Sum(nil))
}

func IsUUID(input string) bool {
	// Regular expression to match UUID format
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(input)
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
