package common

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type tableInfo struct {
	TableName   string       `json:"table_name"`
	IsView      bool         `json:"is_view"`
	TotalRow    int          `json:"total_row"`
	TotalColumn int          `json:"total_column"`
	ColumnList  []columnInfo `json:"column_list,omitempty"` // Use omitempty to skip for views
}

type columnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func FiberTableInfo(app *fiber.App) {
	app.Get("/tables", func(c *fiber.Ctx) error {
		var tables []struct {
			TableName string `gorm:"column:Tables_in_auction-dev"`
			TableType string `gorm:"column:Table_type"`
		}
		if result := Database.Raw("SHOW FULL TABLES").Scan(&tables); result.Error != nil {
			return result.Error
		}

		var tableInfos []tableInfo
		for _, table := range tables {
			if table.TableType == "BASE TABLE" { // Check if it's a table
				tableInfos = append(tableInfos, getTableInfo(Database, table.TableName, false))
			}
		}

		return c.JSON(tableInfos)
	})
}

func getTableInfo(db *gorm.DB, name string, isView bool) tableInfo {
	var totalColumns int
	var columns []struct{ Field, Type string }
	if !isView {
		if result := db.Raw("SHOW COLUMNS FROM " + name).Scan(&columns); result.Error != nil {
			panic(result.Error)
		}
		totalColumns = len(columns)
	}

	var columnInfos []columnInfo
	for _, column := range columns {
		columnType := column.Type
		// Set a default type if column type is empty
		if columnType == "" {
			columnType = "UNKNOWN"
		}
		columnInfos = append(columnInfos, columnInfo{Name: column.Field, Type: columnType})
	}

	return tableInfo{
		TableName:   name,
		IsView:      isView,
		TotalRow:    -1, // You need to implement a method to fetch the total rows
		TotalColumn: totalColumns,
		ColumnList:  columnInfos,
	}
}
