// Package db отвечает за подключение к PostgreSQL и применение миграций
// схемы базы данных.
package db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/truenuta/gofermart/internal/migrations"
)

// RunMigrations применяет к переданному соединению все ещё не выполненные
// миграции схемы из встроенного набора migrations.FS
func RunMigrations(db *sql.DB) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	defer source.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
