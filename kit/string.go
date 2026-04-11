package kit

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	reNonAlphaNum = regexp.MustCompile("[^a-zA-Z0-9]+")
	reUUID        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func StringToFloat(value string) float64 {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}

func MD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(text, salt string) string {
	h := sha256.New()
	h.Write([]byte(text + salt))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateID(input string) string {
	clean := reNonAlphaNum.ReplaceAllString(input, "")
	clean = strings.ToLower(clean)
	sum := sha1.Sum([]byte(clean))
	return hex.EncodeToString(sum[:8])
}

func IsUUID(input string) bool {
	return reUUID.MatchString(input)
}

func TitleCase(text string) string {
	titleCaser := cases.Title(language.English)
	normalized := strings.Join(strings.Fields(text), " ")
	return titleCaser.String(strings.ToLower(normalized))
}

func JSONPickKeys(jsonStr string, keys ...string) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}
	return result, nil
}
