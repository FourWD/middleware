package common

import "github.com/gofiber/fiber/v2"

func FiberTableInfo(app *fiber.App) {

	type ColumnInfo struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	type TableInfo struct {
		TableName   string       `json:"table_name"`
		TotalColumn int          `json:"total_column"`
		ColumnList  []ColumnInfo `json:"column_list"`
	}

	app.Get("/tables", func(c *fiber.Ctx) error {
		var tables []string
		if result := Database.Raw("SHOW TABLES").Scan(&tables); result.Error != nil {
			return result.Error
		}

		var tableInfos []TableInfo
		for _, table := range tables {
			var columns []string
			if result := Database.Raw("SHOW COLUMNS FROM " + table).Scan(&columns); result.Error != nil {
				return result.Error
			}

			var columnInfos []ColumnInfo
			for _, column := range columns {
				columnInfos = append(columnInfos, ColumnInfo{Name: column})
			}

			tableInfos = append(tableInfos, TableInfo{
				TableName:   table,
				TotalColumn: len(columns),
				ColumnList:  columnInfos,
			})
		}

		return c.JSON(tableInfos)
	})
}
