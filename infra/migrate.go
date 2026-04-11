package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratedatabase "github.com/golang-migrate/migrate/v4/database"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RunMigrations runs all pending database migrations. It is a no-op if cfg.Migration.Enabled is false.
func RunMigrations(cfg CommonConfig) error {
	if !cfg.Migration.Enabled {
		return nil
	}

	dialector := gormmysql.Open(BuildMySQLDSN(cfg.Database))
	if cfg.Database.Driver == DBDriverPostgres {
		dialector = gormpostgres.Open(BuildPostgresDSN(cfg.Database))
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("resolve sql db for migrations: %w", err)
	}
	defer sqlDB.Close()

	var driver migratedatabase.Driver
	switch cfg.Database.Driver {
	case DBDriverPostgres:
		driver, err = migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	default:
		driver, err = migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	}
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	migrationPath, err := resolveMigrationPath(cfg.Migration.Path)
	if err != nil {
		return err
	}
	hasFiles, err := hasMigrationFiles(migrationPath)
	if err != nil {
		return err
	}
	if !hasFiles {
		return nil
	}

	sourceURL := "file://" + filepath.ToSlash(migrationPath)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.Database.Name, driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func hasMigrationFiles(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("read migration path %q: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") || strings.HasSuffix(name, ".sql") {
			return true, nil
		}
	}

	return false, nil
}

func resolveMigrationPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("migration path is empty")
	}

	candidates := []string{path}
	if !filepath.IsAbs(path) {
		if wd, err := os.Getwd(); err == nil {
			candidates = append([]string{filepath.Join(wd, path)}, candidates...)
		}

		if executable, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(executable), path))
		}
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		info, err := os.Stat(absPath)
		if err == nil && info.IsDir() {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("migration path %q not found", path)
}
