package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

type columnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type tableInfo struct {
	TableName   string       `json:"table_name"`
	IsView      bool         `json:"is_view"`
	TotalColumn int          `json:"total_column"`
	ColumnList  []columnInfo `json:"column_list"`
	Md5         string       `json:"md5"`
}

// FiberTableInfo registers GET /api/table — dev-only schema introspection.
// Returns each table's columns with a stable MD5 so clients can detect drift.
func FiberTableInfo(app *fiber.App) {
	app.Get("/api/table", func(c fiber.Ctx) error {
		if infra.AppInfo.Env == "prod" {
			return infra.FiberError(c, "1003", "not allowed in production environment")
		}

		rows, err := DatabaseSql.Query(`SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE,
		CHARACTER_MAXIMUM_LENGTH FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME, COLUMN_NAME`, infra.GetEnv("DB_NAME", ""))
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error executing query")
		}
		defer rows.Close()

		var (
			tables    []tableInfo
			current   tableInfo
			currentID string
		)
		for rows.Next() {
			var tableName, columnName, dataType string
			var length sql.NullInt64
			if err := rows.Scan(&tableName, &columnName, &dataType, &length); err != nil {
				infra.AppLog.EventError(err, "TABLE_INFO_SCAN_FAILURE", nil, "",
					infra.WithComponent(infra.ComponentDB),
					infra.WithOperation("scan_row"),
					infra.WithLogKind(infra.LogKindError),
					infra.WithField("table", "information_schema.columns"))
				continue
			}

			if tableName != currentID {
				if currentID != "" {
					tables = append(tables, finalizeTable(current))
				}
				current = tableInfo{TableName: tableName, ColumnList: []columnInfo{}}
				currentID = tableName
			}

			current.ColumnList = append(current.ColumnList, columnInfo{
				Name: columnName,
				Type: formatDataType(dataType, length),
			})
		}
		if currentID != "" {
			tables = append(tables, finalizeTable(current))
		}

		jsonData, err := json.Marshal(tables)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON")
		}
		return c.Status(http.StatusOK).Send(jsonData)
	})
}

// finalizeTable fills in TotalColumn and computes the MD5 of the table
// snapshot — used at both the "row transitions to new table" point and at
// end-of-rows.
func finalizeTable(t tableInfo) tableInfo {
	t.TotalColumn = len(t.ColumnList)
	snapshot, _ := json.Marshal(t)
	t.Md5 = kit.MD5(string(snapshot))
	return t
}

func formatDataType(dataType string, length sql.NullInt64) string {
	if length.Valid {
		return fmt.Sprintf("%s (%d)", dataType, length.Int64)
	}
	return dataType
}
