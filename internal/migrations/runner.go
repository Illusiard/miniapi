package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Runner struct {
	migrationsPath string
	dbURL          string
}

func New(migrationsPath, dbURL string) *Runner {
	return &Runner{
		migrationsPath: migrationsPath,
		dbURL:          dbURL,
	}
}

func (r *Runner) Up() error {
	db, err := sql.Open("pgx", r.dbURL)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate postgres driver: %w", err)
	}

	srcURL := "file://" + filepath.ToSlash(r.migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(srcURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		v, dirty, verr := m.Version()
		if verr == nil {
			if dirty {
				return fmt.Errorf("migrate up failed (version=%d, dirty): %w", v, err)
			}
			return fmt.Errorf("migrate up failed (version=%d): %w", v, err)
		}
	}

	return nil
}
