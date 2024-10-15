package common

import (
	"regexp"
	"strings"
)

func IsThaiMobileNumber(mobile string) bool {
	mobile = strings.ReplaceAll(mobile, " ", "")
	mobile = strings.ReplaceAll(mobile, "-", "")

	if strings.HasPrefix(mobile, "+66") {
		mobile = "0" + mobile[3:]
	}
	// Regular expression for Thai mobile number (starts with 06, 08, or 09 and is followed by 8 digits)
	re := regexp.MustCompile(`^0[689]\d{8}$`)
	return re.MatchString(mobile)
}
