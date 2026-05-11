package infra

import (
	"regexp"
	"strconv"
	"strings"
)

// Imperative-style field validators for HTTP handlers — call one at a
// time, return on first failure. Complement to the ValidationCollector
// in validation.go (which collects all errors before returning); use
// these when the handler logic just needs guard rails on a field or two
// before reaching the use-case layer.
//
// Every helper returns nil on success or a *AppError on failure so the
// caller can `return WriteError(c, err.Status, err)` (or surface it
// through whatever envelope shape the project uses) without extra
// wrapping.

// ValidateLen rejects strings longer than max runes (NOT bytes — so a
// 40-char Thai field works correctly under UTF-8). Empty strings are
// accepted; pair with ValidateNonEmpty when both checks are needed.
//
// `field` becomes the error code suffix: INVALID_<UPPER(field)>.
func ValidateLen(field, value string, max int) *AppError {
	if len([]rune(value)) > max {
		return ErrBadRequest(
			"INVALID_"+strings.ToUpper(field),
			field+" must be at most "+strconv.Itoa(max)+" characters",
		)
	}
	return nil
}

// ValidateNonEmpty rejects strings that are whitespace-only or empty.
func ValidateNonEmpty(field, value string) *AppError {
	if strings.TrimSpace(value) == "" {
		return ErrBadRequest(
			"INVALID_"+strings.ToUpper(field),
			field+" is required",
		)
	}
	return nil
}

// simpleEmailRE is intentionally permissive — full RFC 5321/5322
// compliance is pointless for an opt-in admin form. The check rejects
// things that obviously aren't an address (no @, no dot, whitespace,
// etc.) and lets the SMTP layer make the final call.
var simpleEmailRE = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// ValidateEmail returns nil for syntactically valid addresses up to 255
// characters. The address is NOT lower-cased here — caller decides
// (login flows usually do, audit flows don't).
func ValidateEmail(value string) *AppError {
	if !simpleEmailRE.MatchString(value) {
		return ErrBadRequest("INVALID_EMAIL", "email is not a valid address")
	}
	if len(value) > 255 {
		return ErrBadRequest("INVALID_EMAIL", "email must be at most 255 characters")
	}
	return nil
}

// roleNameRE limits role names to an ASCII subset safe to interpolate
// into URL paths, audit-log payloads, and JSON without escaping
// considerations. Built-in names like "Admin"/"User" match this pattern.
var roleNameRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _\-]{0,39}$`)

// ValidateRoleName enforces the role-name pattern (start with letter,
// 1-40 chars total, letters/digits/spaces/'-'/'_' only).
func ValidateRoleName(value string) *AppError {
	if !roleNameRE.MatchString(value) {
		return ErrBadRequest(
			"INVALID_NAME",
			"name must start with a letter and contain only letters, numbers, spaces, '-', '_' (1..40 chars)",
		)
	}
	return nil
}

// ValidatePort returns nil when p is in the TCP/UDP valid range. Used
// by service-config validation across projects that accept a host:port
// pair.
func ValidatePort(p int) *AppError {
	if p < 1 || p > 65535 {
		return ErrBadRequest("INVALID_PORT", "port must be 1..65535")
	}
	return nil
}
