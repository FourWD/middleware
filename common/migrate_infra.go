package common

import (
	"context"
	"fmt"

	"github.com/FourWD/middleware/infra"
)

// MigrateInfra wires infra-provided clients into the package-level globals
// still read by legacy common code.
//
// Populated globals:
//   - common.Database, common.DatabaseSql            (from deps.Data.Databases.Primary)
//   - common.DatabaseMongo, common.DatabaseMongoMiddleware (from deps.Data.Mongo)
//
// Firebase clients (FirestoreClient, AuthClient, FirebaseMessageClient, FirebaseCtx)
// now live in the infra package and are populated by infra.NewApp itself —
// this function does not touch them. Use infra.FirestoreClient, etc. directly.
//
// Call once from the project's Register function so older code keeps working
// while new code migrates to deps-based dependency injection.
func MigrateInfra(deps infra.AppDeps) error {
	if err := bindDatabase(deps); err != nil {
		return err
	}
	bindMongo(deps)
	return nil
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

func bindMongo(deps infra.AppDeps) {
	mc := deps.Data.Mongo
	if mc == nil {
		return
	}
	wrapped := &MongoDB{
		Client:   mc.Client(),
		Database: mc.Database(),
		Ctx:      context.Background(),
	}
	DatabaseMongo = wrapped
	DatabaseMongoMiddleware = wrapped
}
