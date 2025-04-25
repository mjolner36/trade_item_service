package migrator

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
)

func main() {
	var storagePath, migrationPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "Path to storage file")
	flag.StringVar(&migrationPath, "migration-path", "", "Path to migration file")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migration table")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}

	if migrationPath == "" {
		panic("migration-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		fmt.Sprintf("file://%s", storagePath, migrationsTable))
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}
