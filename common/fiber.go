package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func FiberDisableXFrame(c *fiber.Ctx) error {
	c.Set("X-Frame-Options", "DENY")
	return c.Next()
}

func FiberNoSniff(c *fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Next()
}

/* func FiberError(c *fiber.Ctx, errorCode ...string) error {
	if errorCode[0] != "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 0, "code": errorCode[0], "message": "error"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": 0, "message": "error"})
} */

// func FiberCustom(c *fiber.Ctx, status int, errorCode string, errorMessage string) error {
// 	responseLog(c)
// 	return c.Status(status).JSON(fiber.Map{"status": status, "code": errorCode, "message": errorMessage})
// }

func FiberCustom(c *fiber.Ctx, HTTPStatus int, data map[string]interface{}) error {
	responseLog(c)
	c.Set("Content-Type", "application/json")
	return c.Status(HTTPStatus).JSON(data)
}

func FiberOK(c *fiber.Ctx, status int, code string, message string) error {
	response := map[string]interface{}{
		"status":  status,
		"code":    code,
		"message": message,
	}
	return FiberCustom(c, fiber.StatusOK, response)
}

func FiberSuccess(c *fiber.Ctx) error {
	// responseLog(c)
	// return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": 1, "message": "success"})

	// response := map[string]interface{}{
	// 	"status":  1,
	// 	"message": "success",
	// }
	// return FiberCustom(c, fiber.StatusOK, response)
	return FiberOK(c, 1, "0000", "success")
}

func FiberError(c *fiber.Ctx, code string, message string, err ...error) error {
	response := map[string]interface{}{
		"status":  0,
		"code":    code,
		"message": message,
	}
	return FiberCustom(c, fiber.StatusInternalServerError, response)
}

func FiberReviewPayload(c *fiber.Ctx) error {
	return FiberError(c, "1002", "review your payload")
}

// func FiberErrorSql(c *fiber.Ctx, errorMessage string) error {
// 	log.Println("Error", errorMessage)
// 	return FiberCustom(c, fiber.StatusInternalServerError, "500", "SQL ERROR")
// }

// func FiberErrorFirebase(c *fiber.Ctx, errorMessage string) error {
// 	log.Println("Error", errorMessage)
// 	return FiberCustom(c, fiber.StatusInternalServerError, "501", "FIREBASE ERROR")
// }

func FiberQueryWithCustomDB(c *fiber.Ctx, db *sql.DB, sql string, values ...interface{}) error {
	jsonBytes, sql, err := queryToJSON(db, sql, values...)
	if err != nil {
		// PrintError(`SQL Error`, err.Error())
		return FiberError(c, "1001", "sql error")
	}
	return FiberSendData(c, string(jsonBytes), sql)
}

func FiberQuery(c *fiber.Ctx, sql string, values ...interface{}) error {
	return FiberQueryWithCustomDB(c, DatabaseMysql, sql, values...)
}

func FiberQueryWithCustomDBLimit1(c *fiber.Ctx, db *sql.DB, sql string, values ...interface{}) error {
	jsonBytes, sql, err := queryToJSON(db, sql, values...)
	var result []map[string]interface{}
	if json.Unmarshal(jsonBytes, &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return FiberError(c, "1001", "sql error")
	}

	if len(result) == 1 {
		firstRow := result[0]
		if firstRowJSON, err := json.Marshal(firstRow); err == nil {
			return FiberSendData(c, string(firstRowJSON), sql)
		}
	}

	return FiberSendData(c, "", "")
}

func FiberQueryLimit1(c *fiber.Ctx, sql string, values ...interface{}) error {
	return FiberQueryWithCustomDBLimit1(c, DatabaseMysql, sql, values...)
}

func FiberSendData(c *fiber.Ctx, jsonData string, sql string) error {
	response := map[string]interface{}{
		"status":  1,
		"message": "success",
		"data":    json.RawMessage(jsonData), // Ensures `jsonData` is treated as raw JSON
	}

	if App.Env != "prod" {
		response["sql"] = sql
	}

	// responseLog(c)
	// return c.JSON(response)
	return FiberCustom(c, fiber.StatusOK, response)
}

func FiberDeleteByID(c *fiber.Ctx, tableName string) error {
	type Delete struct {
		ID       string `json:"id"`
		DeleteBy string `json:"delete_by"`
	}

	var payload Delete
	err := c.BodyParser(payload)
	if err != nil {
		return FiberReviewPayload(c)
	}

	if tableName == `` || payload.ID == `` || payload.DeleteBy == `` {
		return FiberReviewPayload(c)
	}

	result := Database.Exec(`UPDATE ? SET deleted_at = now(), deleted_by = ? WHERE id = ?`, tableName, payload.DeleteBy, payload.ID)
	if result.Error != nil {
		// PrintError(`FiberDelete`, result.Error.Error())
		return FiberError(c, "1001", "sql error")
	} //fmt.Println("Affected Rows:", result.RowsAffected)

	return FiberSuccess(c)
}

func FiberDeletePermanentByID(c *fiber.Ctx, tableName string) error {
	type Delete struct {
		ID       string `json:"id"`
		DeleteBy string `json:"delete_by"`
	}

	var payload Delete
	err := c.BodyParser(payload)
	if err != nil {
		return FiberReviewPayload(c)
	}

	if tableName == `` || payload.ID == `` {
		return FiberReviewPayload(c)
	}

	result := Database.Exec(`DELETE FROM ? WHERE id = ?`, tableName, payload.ID)
	if result.Error != nil {
		// PrintError(`FiberDeletePermanent`, result.Error.Error())
		return FiberError(c, "1001", "sql error")
	}

	return FiberSuccess(c)
}

func FiberWarmUp(app *fiber.App) {
	app.Get("/_ah/warmup", func(c *fiber.Ctx) error {
		// message := "Warm-up request succeeded"
		// responseLog(c)
		// jsonData := `{"message":"Warm-up request succeeded"}`
		// c.Set("Content-Type", "application/json")
		// return c.Status(http.StatusOK).SendString(jsonData)
		// return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Warm-up request succeeded"})

		// response := map[string]interface{}{
		// 	"status":  1,
		// 	"message": "Warm-up request succeeded",
		// }

		// return FiberCustom(c, fiber.StatusOK, response)
		return FiberOK(c, 1, "0000", "Warm-up request succeeded")
	})
}

func FiberSql(app *fiber.App) {
	type Payload struct {
		Query string `json:"query"`
	}

	app.Post("/sql", func(c *fiber.Ctx) error {
		payload := new(Payload)
		if err := c.BodyParser(payload); err != nil {
			return FiberReviewPayload(c)
		}
		return FiberQuery(c, payload.Query)
	})
}

func FiberWakeUp(app *fiber.App) {
	app.Get("/wake-up", func(c *fiber.Ctx) error {
		App.AppVersion = viper.GetString("app_version")
		// jsonData := `{"status":1, "message":"success", "data":` + StructToString(App) + `}`
		// c.Set("Content-Type", "application/json")
		// return c.SendString(string(jsonData))
		//return c.Status(http.StatusOK).JSON(fiber.Map{"status": 1, "message": "success", "data": StructToString(App)})

		// return c.Status(http.StatusOK).JSON(fiber.Map{"status": 1, "message": "success", "data": App})

		response := map[string]interface{}{
			"status":  1,
			"message": "success",
			"data":    App,
		}

		return FiberCustom(c, fiber.StatusOK, response)
	})
}

func rawSql(sql string, values ...interface{}) string {
	parts := strings.Split(sql, "?")
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

	full := strings.ReplaceAll(fullSQL.String(), "\n", " ") // Remove newlines
	full = strings.ReplaceAll(full, "\t", " ")              // Remove tabs
	full = strings.ReplaceAll(full, `\"`, "")
	full = strings.TrimSpace(full)

	return full
}

func queryToJSON(db *sql.DB, sql string, values ...interface{}) ([]byte, string, error) {
	list := []string{"INSERT ", "UPDATE ", "DELETE ", "CREATE ", "EMPTY ", "DROP ", "ALTER ", "TRUNCATE "}
	if StringExistsInList(strings.ToUpper(sql), list) {
		return nil, "", errors.New("NOT ALLOW: INSERT/UPDATE/DELETE/CREATE/EMPTY/DROP/ALTER/TRUNCATE")
	}

	// Log the SQL query and values for debugging
	// log.Printf("Executing SQL: %s, with values: %v", sql, values)

	rows, err := db.Query(sql, values...)
	// log.Println(sql)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, "", err
	}

	// types, err := rows.ColumnTypes()
	// if err != nil {
	// 	return nil, err
	// }

	result := make([]map[string]interface{}, 0)
	//result := make([]map[string]string, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, "", err
		}

		m := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}

			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				if val != nil {
					temp := fmt.Sprintf("%v", val)
					temp = strings.Replace(temp, " +0700 +07", "", -1)
					v = temp
					if len(temp) >= 10 {
						if temp[0:10] == "1900-01-01" {
							v = nil
						}
					}
				} else {
					v = val
				}
			}
			m[col] = v
		}
		result = append(result, m)
	}

	raw := ""
	if App.Env != "prod" {
		raw = rawSql(sql, values...)
	}
	jByte, err := json.Marshal(result)

	return jByte, raw, err
}
