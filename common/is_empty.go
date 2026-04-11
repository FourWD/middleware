package common

import (
	"fmt"
	"strings"
)

func IsEmpty(params map[string]string) error { // Check if any field is empty
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
