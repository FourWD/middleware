package common

import (
	"regexp"
)

func IsThaiMobileNumber(mobile string) bool {
	// Regular expression for Thai mobile number that starts with +66 or 0
	re := regexp.MustCompile(`^(\+66|0)[689]\d{8}$`)
	return re.MatchString(mobile)
}
