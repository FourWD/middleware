package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

func FiberPaginatedQuery(c fiber.Ctx, baseSQL string, values ...interface{}) error {
	page, limit, offset := paginationParams(c)

	if !kit.IsReadOnlySQL(baseSQL) {
		return infra.FiberError(c, "1001", "NOT ALLOW: only SELECT/WITH/SHOW/EXPLAIN/DESC statements are permitted")
	}

	paginatedSQL := fmt.Sprintf(
		"SELECT *, count(*) OVER() AS full_count FROM (%s) AS sub LIMIT %d OFFSET %d",
		baseSQL, limit, offset,
	)

	dialect := infra.SQLDialect(DatabaseSql)

	stmt := paginatedSQL
	if dialect == infra.DBDriverPostgres {
		stmt = kit.ToPostgresPlaceholders(paginatedSQL)
	}

	rows, err := DatabaseSql.Query(stmt, values...)
	if err != nil {
		return infra.FiberError(c, "1001", "sql error")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return infra.FiberError(c, "1001", "column read error")
	}
	dbTypes := infra.SQLColumnTypes(rows, len(columns), dialect)

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
			v := infra.ConvertSQLValue(rowVals[i], dialect, infra.SQLColumnTypeAt(dbTypes, i))
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
