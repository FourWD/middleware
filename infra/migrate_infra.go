package infra

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// Database and DatabaseSql are package-level handles populated by MigrateInfra.
// They exist so legacy code can reach the primary DB without carrying AppDeps.
// Prefer passing deps.Data.Databases explicitly where possible.
var (
	Database    *gorm.DB
	DatabaseSql *sql.DB
)

// databaseSyncHook allows packages outside infra (e.g. common) to mirror the
// Database/DatabaseSql values into their own globals without introducing a
// circular import. Register via RegisterDatabaseSync.
var databaseSyncHook func(*gorm.DB, *sql.DB)

// RegisterDatabaseSync registers a callback invoked by MigrateInfra after the
// primary DB handles are resolved. Use this from common or other legacy
// packages that expose their own Database vars for backwards compatibility.
func RegisterDatabaseSync(hook func(*gorm.DB, *sql.DB)) {
	databaseSyncHook = hook
}

// MigrateInfra binds the primary database from deps into the infra-level
// Database/DatabaseSql globals and invokes any registered sync hook.
// NewApp calls this automatically — downstream code no longer needs to.
func MigrateInfra(deps AppDeps) error {
	if deps.Data.Databases.Primary == nil {
		return nil
	}
	db, sqlDB, err := BindDatabase(deps.Data.Databases)
	if err != nil {
		return fmt.Errorf("migrate infra: %w", err)
	}
	Database = db
	DatabaseSql = sqlDB
	if databaseSyncHook != nil {
		databaseSyncHook(db, sqlDB)
	}
	return nil
}
