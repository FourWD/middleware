package infra

import (
	"database/sql"
	"fmt"
	"time"
)

func ScanRows(rows *sql.Rows) ([]map[string]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	result := make([]map[string]string, 0)
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]string, len(columns))
		for i, colName := range columns {
			rowMap[colName] = rowString(values[i])
		}
		result = append(result, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func rowString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(v)
	}
}

func AsString(v string) string { return v }

func AsInt(v string) int {
	var n int
	fmt.Sscanf(v, "%d", &n)
	return n
}

func AsTime(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", v)
	}
	if err != nil {
		return time.Time{}
	}
	return t
}
