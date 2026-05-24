package kit

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

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

func StringExistsInList(target string, strList []string) bool {
	return slices.Contains(strList, target)
}

// MaskMobile returns a log-safe representation of a mobile number: the last
// four digits prefixed with "XXX-XXX-". Inputs shorter than 4 chars become
// "XXXX" so callers never accidentally log a full number.
func MaskMobile(mobile string) string {
	if len(mobile) < 4 {
		return "XXXX"
	}
	return "XXX-XXX-" + mobile[len(mobile)-4:]
}

// MaskEmail returns just the domain portion ("@example.com" → "example.com");
// the local part is replaced with "***" to stay PII-safe. Inputs without
// "@" return "***".
func MaskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at < 0 || at == len(email)-1 {
		return "***"
	}
	return "***@" + email[at+1:]
}

func DateString() string {
	var b strings.Builder
	b.Grow(19 + 10)
	b.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
	appendRandomDigits(&b, 10)
	return b.String()
}

func generateRandomDigits(count int) string {
	var b strings.Builder
	b.Grow(count)
	appendRandomDigits(&b, count)
	return b.String()
}

func appendRandomDigits(b *strings.Builder, count int) {
	for i := 0; i < count; i++ {
		b.WriteByte('0' + byte(rand.IntN(10)))
	}
}

func DateToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DateFormatSecond)
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
