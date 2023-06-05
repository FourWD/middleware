package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/FourWD/middleware/payload"
	"github.com/gofiber/fiber/v2"
)

func FiberReviewPayload(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 0, "message": "review your payload"})
}

func FiberSuccess(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": 1, "message": "success"})
}

func FiberError(c *fiber.Ctx, errorCode ...string) error {
	if errorCode[0] != "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 0, "code": errorCode[0], "message": "error"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 0, "message": "error"})
}

func FiberQuery(c *fiber.Ctx, sql string) error {
	jsonBytes, err := QueryToJSON(DatabaseMysql, sql)
	if err != nil {
		PrintError(`SQL Error`, err.Error())
		return FiberError(c)
	}
	return FiberSendData(c, string(jsonBytes))
}

func FiberSendData(c *fiber.Ctx, json string) error {
	message := `{"status":1, "message":"success", "data": ` + json + `}`
	c.Set("Content-Type", "application/json")
	return c.SendString(string(message))
}

func FiberDeleteByID(c *fiber.Ctx, tableName string) error {
	var payload payload.Delete
	err := c.BodyParser(payload)
	if err != nil {
		return FiberReviewPayload(c)
	}

	if tableName == `` || payload.ID == `` || payload.DeleteBy == `` {
		return FiberReviewPayload(c)
	}

	result := Database.Exec(`UPDATE ? SET deleted_at = now(), deleted_by = ? WHERE id = ?`, tableName, payload.DeleteBy, payload.ID)
	if result.Error != nil {
		PrintError(`FiberDelete`, result.Error.Error())
		return FiberError(c)
	} //fmt.Println("Affected Rows:", result.RowsAffected)

	return FiberSuccess(c)
}

func FiberDeletePermanentByID(c *fiber.Ctx, tableName string) error {
	var payload payload.Delete
	err := c.BodyParser(payload)
	if err != nil {
		return FiberReviewPayload(c)
	}

	if tableName == `` || payload.ID == `` {
		return FiberReviewPayload(c)
	}

	result := Database.Exec(`DELETE FROM ? WHERE id = ?`, tableName, payload.ID)
	if result.Error != nil {
		PrintError(`FiberDeletePermanent`, result.Error.Error())
		return FiberError(c)
	}

	return FiberSuccess(c)
}

func QueryToJSON(db *sql.DB, query string) ([]byte, error) {
	list := []string{"INSERT ", "UPDATE ", "DELETE ", "CREATE ", "EMPTY ", "DROP ", "ALTER ", "TRUNCATE "}
	if containsAny(strings.ToUpper(query), list) {
		return nil, errors.New("NOT ALLOW: INSERT/UPDATE/DELETE/CREATE/EMPTY/DROP/ALTER/TRUNCATE")
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			m[col] = v
		}
		result = append(result, m)
	}

	return json.Marshal(result)
}

func containsAny(target string, list []string) bool {
	for _, str := range list {
		if strings.Contains(target, str) {
			return true
		}
	}
	return false
}
