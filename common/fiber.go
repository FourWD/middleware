package common

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/FourWD/middleware/infra"
	"github.com/gofiber/fiber/v3"
)

// Deprecated: use infra.FiberCustom directly.
func FiberCustom(c fiber.Ctx, HTTPStatus int, data map[string]interface{}) error {
	return infra.FiberCustom(c, HTTPStatus, data)
}

// Deprecated: use infra.FiberOK directly.
func FiberOK(c fiber.Ctx, status int, code string, message string) error {
	return infra.FiberOK(c, status, code, message)
}

// Deprecated: use infra.FiberSuccess directly.
func FiberSuccess(c fiber.Ctx) error {
	return infra.FiberSuccess(c)
}

// Deprecated: use infra.FiberError directly.
func FiberError(c fiber.Ctx, code string, message string, err ...error) error {
	return infra.FiberError(c, code, message, err...)
}

// Deprecated: use infra.FiberReviewPayload directly.
func FiberReviewPayload(c fiber.Ctx) error {
	return infra.FiberReviewPayload(c)
}

// Deprecated: use infra.FiberSendData directly.
func FiberSendData(c fiber.Ctx, jsonData string, sql string) error {
	return infra.FiberSendData(c, jsonData, sql)
}

// Deprecated: use infra.FiberQueryWithCustomDB directly.
func FiberQueryWithCustomDB(c fiber.Ctx, db *sql.DB, sqlText string, values ...interface{}) error {
	return infra.FiberQueryWithCustomDB(c, db, sqlText, values...)
}

// Deprecated: use infra.FiberQueryWithCustomDBLimit1 directly.
func FiberQueryWithCustomDBLimit1(c fiber.Ctx, db *sql.DB, sqlText string, values ...interface{}) error {
	return infra.FiberQueryWithCustomDBLimit1(c, db, sqlText, values...)
}

// FiberQuery runs a read-only SQL against the default DatabaseSql connection.
// Stays in common because DatabaseSql is a common-package global.
func FiberQuery(c fiber.Ctx, sqlText string, values ...interface{}) error {
	return infra.FiberQueryWithCustomDB(c, DatabaseSql, sqlText, values...)
}

// FiberQueryLimit1 runs a read-only SQL against DatabaseSql and returns only
// the first row.
func FiberQueryLimit1(c fiber.Ctx, sqlText string, values ...interface{}) error {
	return infra.FiberQueryWithCustomDBLimit1(c, DatabaseSql, sqlText, values...)
}

// FiberSql registers a /sql endpoint (non-prod only) that executes the
// request body as a raw SQL query against DatabaseSql.
func FiberSql(app *fiber.App) {
	app.Post("/sql", func(c fiber.Ctx) error {
		if infra.AppInfo.Env == "prod" {
			return infra.FiberError(c, "1003", "not allowed in production environment")
		}

		type Payload struct {
			Query string `json:"query"`
		}

		payload := new(Payload)
		if err := c.Bind().Body(payload); err != nil {
			return infra.FiberReviewPayload(c)
		}
		return FiberQuery(c, payload.Query)
	})
}

// rawSql interpolates `?` placeholders for debug output. Kept in common
// because FiberPaginatedQuery still uses it; will be removed when
// FiberPaginatedQuery migrates.
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
