package common

import (
	"fmt"
	"strings"
)

func IsEmpty(params map[string]string) error {
	var emptyFields []string

	// Loop through the map and check if any value is empty
	for key, value := range params {
		if value == "" {
			emptyFields = append(emptyFields, key)
		}
	}

	// If any fields are empty, return an error with the field names
	if len(emptyFields) > 0 {
		return fmt.Errorf("%s is/are empty", strings.Join(emptyFields, ", "))
	}

	// If all fields are non-empty, return nil
	return nil
}
