package kit

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reEmail      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	reThaiMobile = regexp.MustCompile(`^(\+66|0)[689]\d{8}$`)
)

func IsEmailValid(email string) bool {
	return reEmail.MatchString(email)
}

func IsThaiMobileNumber(mobile string) bool {
	return reThaiMobile.MatchString(mobile)
}

func IsEmpty(params map[string]string) error {
	var emptyFields []string

	for key, value := range params {
		if value == "" {
			emptyFields = append(emptyFields, key)
		}
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%s is/are empty", strings.Join(emptyFields, ", "))
	}

	return nil
}
