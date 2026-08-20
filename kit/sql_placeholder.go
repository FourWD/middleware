package kit

import (
	"strconv"
	"strings"
)

// ToPostgresPlaceholders rewrites MySQL-style `?` placeholders to the
// positional `$1, $2, ...` form Postgres expects.
//
// `?` inside single-quoted literals, quoted identifiers, dollar-quoted bodies
// and comments is left alone. `??` collapses to a literal `?` so the jsonb
// operators (? ?| ?&) can still be written.
func ToPostgresPlaceholders(sqlText string) string {
	var b strings.Builder
	b.Grow(len(sqlText) + 8)

	n := 0
	for i := 0; i < len(sqlText); {
		switch c := sqlText[i]; {
		case c == '\'', c == '"', c == '`':
			j := scanQuoted(sqlText, i, c)
			b.WriteString(sqlText[i:j])
			i = j

		case c == '-' && i+1 < len(sqlText) && sqlText[i+1] == '-':
			j := scanLineComment(sqlText, i)
			b.WriteString(sqlText[i:j])
			i = j

		case c == '/' && i+1 < len(sqlText) && sqlText[i+1] == '*':
			j := scanBlockComment(sqlText, i)
			b.WriteString(sqlText[i:j])
			i = j

		case c == '$':
			j, ok := scanDollarQuoted(sqlText, i)
			if !ok {
				b.WriteByte(c)
				i++
				continue
			}
			b.WriteString(sqlText[i:j])
			i = j

		case c == '?':
			if i+1 < len(sqlText) && sqlText[i+1] == '?' {
				b.WriteByte('?')
				i += 2
				continue
			}
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			i++

		default:
			b.WriteByte(c)
			i++
		}
	}

	return b.String()
}

// scanQuoted returns the index just past the closing quote. A doubled quote is
// an escaped quote, not a terminator. An unterminated literal consumes the rest.
func scanQuoted(s string, start int, quote byte) int {
	for i := start + 1; i < len(s); i++ {
		if s[i] != quote {
			continue
		}
		if i+1 < len(s) && s[i+1] == quote {
			i++
			continue
		}
		return i + 1
	}
	return len(s)
}

func scanLineComment(s string, start int) int {
	if j := strings.IndexByte(s[start:], '\n'); j >= 0 {
		return start + j + 1
	}
	return len(s)
}

// scanBlockComment handles Postgres' nestable /* */ comments.
func scanBlockComment(s string, start int) int {
	depth := 1
	for i := start + 2; i < len(s)-1; {
		switch {
		case s[i] == '/' && s[i+1] == '*':
			depth++
			i += 2
		case s[i] == '*' && s[i+1] == '/':
			depth--
			i += 2
			if depth == 0 {
				return i
			}
		default:
			i++
		}
	}
	return len(s)
}

// scanDollarQuoted matches $tag$...$tag$ (tag may be empty). Reports false for
// anything else at start, including an already-positional $1.
func scanDollarQuoted(s string, start int) (int, bool) {
	end := start + 1
	for end < len(s) && isTagByte(s[end], end == start+1) {
		end++
	}
	if end >= len(s) || s[end] != '$' {
		return 0, false
	}

	tag := s[start : end+1]
	if j := strings.Index(s[end+1:], tag); j >= 0 {
		return end + 1 + j + len(tag), true
	}
	return len(s), true
}

func isTagByte(c byte, first bool) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
		return true
	case !first && c >= '0' && c <= '9':
		return true
	}
	return false
}
