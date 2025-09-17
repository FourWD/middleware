package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FiberPaginatedQuery(c *fiber.Ctx, baseSQL string, values ...interface{}) error {
	// Step 1: Handle pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Step 2: Ensure no unsafe SQL command
	blocked := []string{"INSERT ", "UPDATE ", "DELETE ", "CREATE ", "EMPTY ", "DROP ", "ALTER ", "TRUNCATE "}
	if StringExistsInList(strings.ToUpper(baseSQL), blocked) {
		return FiberError(c, "1001", "NOT ALLOW: INSERT/UPDATE/DELETE/CREATE/EMPTY/DROP/ALTER/TRUNCATE")
	}

	// Step 3: Construct paginated SQL
	paginatedSQL := fmt.Sprintf("SELECT *, count(*) OVER() AS full_count FROM (%s) AS sub LIMIT %d OFFSET %d", baseSQL, limit, offset)

	// Step 4: Execute query
	rows, err := DatabaseSql.Query(paginatedSQL, values...)
	if err != nil {
		return FiberError(c, "1001", "sql error")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return FiberError(c, "1001", "column read error")
	}

	var result []map[string]interface{}
	totalItems := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return FiberError(c, "1001", "row scan error")
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			var v interface{}
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				if val != nil {
					temp := fmt.Sprintf("%v", val)
					temp = strings.Replace(temp, " +0700 +07", "", -1)
					if len(temp) >= 10 && temp[:10] == "1900-01-01" {
						v = nil
					} else {
						v = temp
					}
				} else {
					v = nil
				}
			}
			rowMap[col] = v
			if strings.ToLower(col) == "full_count" {
				if str, ok := v.(string); ok {
					totalItems, _ = strconv.Atoi(str)
				} else if num, ok := v.(int64); ok {
					totalItems = int(num)
				} else if num, ok := v.(int); ok {
					totalItems = num
				}
			}
		}
		result = append(result, rowMap)
	}

	totalPages := (totalItems + limit - 1) / limit
	sqlDebug := ""
	if App.Env != "prod" {
		sqlDebug = rawSql(paginatedSQL, values...)
	}

	// Step 5: Return JSON
	return FiberCustom(c, fiber.StatusOK, fiber.Map{
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
