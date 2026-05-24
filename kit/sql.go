package kit

import (
	"strings"
)

func CheckSqlInjection(text string) string {
	list := []string{"INSERT ", "UPDATE ", "DELETE ", "CREATE ", "EMPTY ", "DROP ", "ALTER ", "TRUNCATE ", "UNION ", ";"}
	if StringExistsInList(strings.ToUpper(text), list) {
		return "ERROR"
	}
	return text
}

var readOnlySQLPrefixes = []string{"SELECT", "WITH", "SHOW", "EXPLAIN", "DESC", "DESCRIBE"}

// IsReadOnlySQL reports whether sql is a read-only statement.
// Allowed prefixes: SELECT / WITH / SHOW / EXPLAIN / DESC / DESCRIBE.
// Stacked statements (a ';' followed by more SQL) are rejected.
func IsReadOnlySQL(sql string) bool {
	trimmed := strings.TrimLeft(sql, " \t\r\n(")
	if trimmed == "" {
		return false
	}

	if i := strings.Index(trimmed, ";"); i >= 0 {
		if strings.TrimSpace(trimmed[i+1:]) != "" {
			return false
		}
	}

	upper := strings.ToUpper(trimmed)
	for _, prefix := range readOnlySQLPrefixes {
		if !strings.HasPrefix(upper, prefix) {
			continue
		}
		rest := upper[len(prefix):]
		if rest == "" {
			return true
		}
		switch rest[0] {
		case ' ', '\t', '\n', '\r', '(':
			return true
		}
	}
	return false
}
