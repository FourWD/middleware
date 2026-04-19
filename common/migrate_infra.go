package common

import (
	"fmt"

	"github.com/FourWD/middleware/infra"
)

// MigrateInfra wires infra-provided clients into the package-level globals
// still read by legacy common code (Database, DatabaseSql).
//
// Firebase, Mongo, AppInfo, AppLog are populated by infra.NewApp itself —
// this function does not touch them. Use infra.FirestoreClient, infra.Mongo,
// infra.AppInfo, infra.AppLog directly.
//
// Call once from the project's Register function so older code keeps working
// while new code migrates to deps-based dependency injection.
func MigrateInfra(deps infra.AppDeps) error {
	return bindDatabase(deps)
}

func bindDatabase(deps infra.AppDeps) error {
	if deps.Data.Databases.Primary == nil {
		return nil
	}
	db, sqlDB, err := infra.BindDatabase(deps.Data.Databases)
	if err != nil {
		return fmt.Errorf("migrate infra: %w", err)
	}
	Database = db
	DatabaseSql = sqlDB
	return nil
}
