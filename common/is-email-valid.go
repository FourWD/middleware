package common

import "regexp"

func IsEmailValid(email string) bool {
	// Define the regular expression for a valid email
	const emailRegex = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

	// Compile the regular expression
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
