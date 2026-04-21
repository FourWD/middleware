package common

import (
	"database/sql"

	"github.com/FourWD/middleware/infra"
	"gorm.io/gorm"
)

// DatabaseMongo mirrors the primary MongoDB client (MONGO_URI) for legacy
// code that reads the common.DatabaseMongo global.
//
// The middleware MongoDB (MONGO_MIDDLEWARE_URI) is intentionally NOT exposed
// here — it is reserved for infra internals like the JWT blacklist and can
// only be reached via infra.MongoMiddleware.
var DatabaseMongo *infra.MongoClient

// init registers sync hooks so NewApp also populates the common-level
// globals that legacy code still reads.
func init() {
	infra.RegisterDatabaseSync(func(db *gorm.DB, sqlDB *sql.DB) {
		Database = db
		DatabaseSql = sqlDB
	})
	infra.RegisterMongoSync(func(mc *infra.MongoClient) {
		DatabaseMongo = mc
	})
}
