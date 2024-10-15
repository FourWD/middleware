package common

import "regexp"

func IsThaiMobileNumber(mobile string) bool {
	// Regular expression for Thai mobile number (starts with 06, 08, or 09 and is followed by 8 digits)
	re := regexp.MustCompile(`^0[689]\d{8}$`)
	return re.MatchString(mobile)
}
