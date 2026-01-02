package common

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func StringToFloat(value string) float64 {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}

func parseWithTimezone(strDateTime string, format string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		LogError("STRING_PARSE_ERROR", map[string]interface{}{"error": err.Error(), "input": strDateTime}, "")
		return NilDate(), err
	}

	parsedTime, err := time.ParseInLocation(format, strDateTime, location)
	if err != nil {
		LogError("STRING_PARSE_ERROR", map[string]interface{}{"error": err.Error(), "input": strDateTime}, "")
		return NilDate(), err
	}

	return parsedTime, nil
}

func StringToDateTime(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_MINUTE)
}

func StringToDate(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_DAY)
}

func MD5(text string) string {
	hashText := md5.New()
	hashText.Write([]byte(text))
	return hex.EncodeToString(hashText.Sum(nil))
}

func Hash(text string, salt string) string {
	hashText := sha256.New()
	hashText.Write([]byte(text + salt))
	return hex.EncodeToString(hashText.Sum(nil))
}

func GenerateID(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	clean := reg.ReplaceAllString(input, "")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ToLower(clean)

	sum := sha1.Sum([]byte(clean))
	return hex.EncodeToString(sum[:8])
}

func IsUUID(input string) bool {
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

func TitleCase(text string) string {
	titleCaser := cases.Title(language.English)
	normalized := strings.Join(strings.Fields(text), " ")
	return titleCaser.String(strings.ToLower(normalized))
}
