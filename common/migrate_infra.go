package common

import (
	"database/sql"

	"github.com/FourWD/middleware/infra"
	"gorm.io/gorm"
)

// init registers a sync hook so infra.MigrateInfra also populates the
// common.Database / common.DatabaseSql globals that legacy code still reads.
func init() {
	infra.RegisterDatabaseSync(func(db *gorm.DB, sqlDB *sql.DB) {
		Database = db
		DatabaseSql = sqlDB
	})
}
