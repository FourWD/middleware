package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

// nullDateMarker is the sentinel "default" date that some legacy MySQL
// columns use in place of NULL. Rows that begin with this date are
// surfaced to API clients as JSON null.
const nullDateMarker = "1900-01-01"

// tzSuffix is the trailing tokens that database/sql appends to TIMESTAMP
// values when the session timezone is Asia/Bangkok — strip them so JSON
// clients see plain "YYYY-MM-DD HH:MM:SS".
const tzSuffix = " +0700 +07"

func FiberPaginatedQuery(c fiber.Ctx, baseSQL string, values ...interface{}) error {
	page, limit, offset := paginationParams(c)

	if !kit.IsReadOnlySQL(baseSQL) {
		return infra.FiberError(c, "1001", "NOT ALLOW: only SELECT/WITH/SHOW/EXPLAIN/DESC statements are permitted")
	}

	paginatedSQL := fmt.Sprintf(
		"SELECT *, count(*) OVER() AS full_count FROM (%s) AS sub LIMIT %d OFFSET %d",
		baseSQL, limit, offset,
	)

	rows, err := DatabaseSql.Query(paginatedSQL, values...)
	if err != nil {
		return infra.FiberError(c, "1001", "sql error")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return infra.FiberError(c, "1001", "column read error")
	}

	var result []map[string]interface{}
	totalItems := 0
	for rows.Next() {
		rowVals := make([]interface{}, len(columns))
		rowPtrs := make([]interface{}, len(columns))
		for i := range columns {
			rowPtrs[i] = &rowVals[i]
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			return infra.FiberError(c, "1001", "row scan error")
		}

		rowMap := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			v := convertSQLValue(rowVals[i])
			rowMap[col] = v
			if strings.ToLower(col) == "full_count" {
				totalItems = parseFullCount(v)
			}
		}
		result = append(result, rowMap)
	}

	totalPages := (totalItems + limit - 1) / limit
	sqlDebug := ""
	if infra.AppInfo.Env != "prod" {
		sqlDebug = rawSql(paginatedSQL, values...)
	}

	return infra.FiberCustom(c, fiber.StatusOK, fiber.Map{
		"status":  1,
		"message": "success",
		"data":    result,
		"pagination": fiber.Map{
			"currentPage": page,
			"pageSize":    limit,
			"totalItems":  totalItems,
			"totalPages":  totalPages,
		},
		"sql": sqlDebug,
	})
}

func paginationParams(c fiber.Ctx) (page, limit, offset int) {
	page = fiber.Query[int](c, "page", 1)
	limit = fiber.Query[int](c, "limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return page, limit, (page - 1) * limit
}

// convertSQLValue normalises raw driver values for JSON output:
// []byte → string, nil → nil, formatted time strings → trimmed string
// or nil when they match the legacy "1900-01-01" sentinel.
func convertSQLValue(val interface{}) interface{} {
	if b, ok := val.([]byte); ok {
		return string(b)
	}
	if val == nil {
		return nil
	}
	s := strings.ReplaceAll(fmt.Sprintf("%v", val), tzSuffix, "")
	if strings.HasPrefix(s, nullDateMarker) {
		return nil
	}
	return s
}

func parseFullCount(v interface{}) int {
	switch n := v.(type) {
	case string:
		i, _ := strconv.Atoi(n)
		return i
	case int64:
		return int(n)
	case int:
		return n
	}
	return 0
}
