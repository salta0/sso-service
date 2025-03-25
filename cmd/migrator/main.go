package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "migrations table name")
	flag.Parse()

	if err := validateRequired(storagePath, migrationsPath, migrationsTable); err != nil {
		panic(err.Error())
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no new migrations to run")

			return
		}
		panic(err)
	}

	fmt.Println("all migrations applied successfully")
}

func validateRequired(storagePath, migrationsPath, migrationsTable string) error {
	if storagePath == "" {
		return errors.New("storage path is required")
	}
	if migrationsPath == "" {
		return errors.New("migrations path is required")
	}
	if migrationsTable == "" {
		return errors.New("migrations table is required")
	}

	return nil
}
