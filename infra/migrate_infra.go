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

// mongoSyncHook mirrors the primary MongoClient into legacy packages that
// expose their own global (e.g. common.DatabaseMongo). Deliberately receives
// only the primary Mongo — the middleware Mongo is reserved for infra
// internals (e.g. the JWT blacklist) and must not leak to common.
var mongoSyncHook func(*MongoClient)

// RegisterMongoSync registers a callback invoked by NewApp after the primary
// Mongo client is resolved. The middleware Mongo is NOT propagated.
func RegisterMongoSync(hook func(*MongoClient)) {
	mongoSyncHook = hook
}

// MigrateInfra binds the primary database and the primary MongoDB into the
// package-level globals (infra.Database, infra.DatabaseSql) and notifies the
// registered sync hooks so that legacy packages such as common can mirror the
// values into their own globals (common.Database, common.DatabaseSql,
// common.DatabaseMongo).
//
// NewApp does NOT call this automatically — projects that still rely on the
// common.* globals must call it from their Register function:
//
//	func Register(app *fiber.App, deps infra.AppDeps) error {
//	    if err := infra.MigrateInfra(deps); err != nil {
//	        return err
//	    }
//	    ...
//	}
//
// New projects that access the database exclusively through deps (or the
// infra.* globals populated by NewApp) can skip this call entirely.
//
// The middleware MongoDB (MONGO_MIDDLEWARE_URI) is intentionally NOT exposed
// through any sync hook — it stays reachable only via infra.MongoMiddleware.
func MigrateInfra(deps AppDeps) error {
	if deps.Data.Databases.Primary != nil {
		db, sqlDB, err := BindDatabase(deps.Data.Databases)
		if err != nil {
			return fmt.Errorf("migrate infra: %w", err)
		}
		Database = db
		DatabaseSql = sqlDB
		if databaseSyncHook != nil {
			databaseSyncHook(db, sqlDB)
		}
	}

	if mongoSyncHook != nil && deps.Data.Mongo != nil {
		mongoSyncHook(deps.Data.Mongo)
	}

	return nil
}
