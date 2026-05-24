package infra

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

// FiberQueryWithCustomDB runs the read-only SQL on the given db and writes
// the result rows as a JSON array via FiberSendData.
func FiberQueryWithCustomDB(c fiber.Ctx, db *sql.DB, sqlText string, values ...interface{}) error {
	jsonBytes, sqlDebug, err := queryToJSON(db, sqlText, values...)
	if err != nil {
		return FiberError(c, "1001", "sql error", err)
	}
	return FiberSendData(c, string(jsonBytes), sqlDebug)
}

// FiberQueryWithCustomDBLimit1 is FiberQueryWithCustomDB but returns only
// the first row as an object (rather than a single-element array).
func FiberQueryWithCustomDBLimit1(c fiber.Ctx, db *sql.DB, sqlText string, values ...interface{}) error {
	jsonBytes, sqlDebug, err := queryToJSON(db, sqlText, values...)
	if err != nil {
		return FiberError(c, "1001", "sql error", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return FiberError(c, "1001", "sql error", err)
	}

	if len(result) == 1 {
		if firstRowJSON, err := json.Marshal(result[0]); err == nil {
			return FiberSendData(c, string(firstRowJSON), sqlDebug)
		}
	}

	return FiberSendData(c, "", "")
}

// rawSql interpolates `?` placeholders with quoted Go values for debug
// output. NOT safe for execution — for log/display only.
func rawSql(sqlText string, values ...interface{}) string {
	parts := strings.Split(sqlText, "?")
	if len(parts)-1 != len(values) {
		return "SQL and values count mismatch"
	}

	var fullSQL strings.Builder
	for i, part := range parts {
		fullSQL.WriteString(part)
		if i < len(values) {
			fullSQL.WriteString(fmt.Sprintf("'%v'", values[i]))
		}
	}

	full := strings.ReplaceAll(fullSQL.String(), "\n", " ")
	full = strings.ReplaceAll(full, "\t", " ")
	full = strings.ReplaceAll(full, `\"`, "")
	return strings.TrimSpace(full)
}

func queryToJSON(db *sql.DB, sqlText string, values ...interface{}) ([]byte, string, error) {
	if !kit.IsReadOnlySQL(sqlText) {
		return nil, "", errors.New("NOT ALLOW: only SELECT/WITH/SHOW/EXPLAIN/DESC statements are permitted")
	}

	rows, err := db.Query(sqlText, values...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, "", err
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		rowVals := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &rowVals[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, "", err
		}

		m := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := rowVals[i]
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else if val != nil {
				temp := fmt.Sprintf("%v", val)
				temp = strings.Replace(temp, " +0700 +07", "", -1)
				v = temp
				if len(temp) >= 10 && temp[0:10] == "1900-01-01" {
					v = nil
				}
			} else {
				v = val
			}
			m[col] = v
		}
		result = append(result, m)
	}

	raw := ""
	if AppInfo.Env != "prod" {
		raw = rawSql(sqlText, values...)
	}
	jByte, err := json.Marshal(result)
	return jByte, raw, err
}
