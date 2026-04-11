package infra

import (
	"errors"
	"strings"
)

var (
	ErrMissingAuthorization = errors.New("authorization header is required")
	ErrInvalidAuthorization = errors.New("authorization header must be Bearer token")
)

func ParseBearerToken(header string) (string, error) {
	if header == "" {
		return "", ErrMissingAuthorization
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidAuthorization
	}

	return parts[1], nil
}
