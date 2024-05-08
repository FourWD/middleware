package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func FiberTableInfo(app *fiber.App) {
	type ColumnInfo struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	type TableInfo struct {
		TableName   string       `json:"table_name"`
		IsView      bool         `json:"is_view"`
		TotalColumn int          `json:"total_column"`
		ColumnList  []ColumnInfo `json:"column_list"`
		Md5         string       `json:"md5"`
	}

	app.Get("/api/v1/tables", func(c *fiber.Ctx) error {
		DBName := viper.GetString("database.database")
		rows, err := DatabaseMysql.Query(`SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, 
		CHARACTER_MAXIMUM_LENGTH FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? 
		ORDER BY TABLE_NAME, COLUMN_NAME`, DBName)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error executing query")
		}
		defer rows.Close()

		// Parse query result and structure JSON response
		var tables []TableInfo
		var currentTable string
		var tableInfo TableInfo
		for rows.Next() {
			var tableName, columnName, dataType string
			var length sql.NullInt64
			err := rows.Scan(&tableName, &columnName, &dataType, &length)
			if err != nil {
				log.Println("Error scanning row:", err)
				continue
			}
			if tableName != currentTable {
				if currentTable != "" {
					tableInfo.TotalColumn = len(tableInfo.ColumnList)
					md5Byte, _ := json.Marshal(tableInfo)
					tableInfo.Md5 = MD5(string(md5Byte))
					tables = append(tables, tableInfo)
				}
				tableInfo = TableInfo{
					TableName:  tableName,
					ColumnList: make([]ColumnInfo, 0),
				}
				currentTable = tableName
			}

			var fullDataType string
			if length.Valid {
				fullDataType = fmt.Sprintf("%s (%d)", dataType, length.Int64)
			} else {
				fullDataType = dataType
			}

			tableInfo.ColumnList = append(tableInfo.ColumnList, ColumnInfo{
				Name: columnName,
				Type: fullDataType,
			})
		}
		if currentTable != "" {
			tableInfo.TotalColumn = len(tableInfo.ColumnList)
			md5Byte, _ := json.Marshal(tableInfo)
			tableInfo.Md5 = MD5(string(md5Byte))
			tables = append(tables, tableInfo)
		}

		// Marshal tables slice to JSON
		jsonData, err := json.Marshal(tables)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON")
		}

		// Return JSON response
		return c.Status(http.StatusOK).Send(jsonData)
	})
}
