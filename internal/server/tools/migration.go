package tools

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"

	// Register packages for migration
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migration(source, dsn string) {
	m, err := migrate.New(source, dsn)
	if err != nil {
		log.Printf("database migration: %s", err)

		return
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("database migration: %s", err)
	}
}
