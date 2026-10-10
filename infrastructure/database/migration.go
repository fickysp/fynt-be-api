package database

import (
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		slog.Error("Failed to init database migration", "error", err)
		os.Exit(1)
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("Failed to execute migrations file", "error", err)
		os.Exit(1)
	}

	slog.Info("database migration success to build!")
}
