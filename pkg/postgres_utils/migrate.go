package postgres_utils

import (
	"errors"
	"fmt"
	"os"

	// lib for the migrations
	"github.com/golang-migrate/migrate/v4"
	// postgres driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// driver for the files
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Action string

const (
	Up   Action = "up"
	Down Action = "down"
)

var (
	ErrMigrationPathEmpty        = errors.New("migration path is empty")
	ErrMigrationActionIsNotValid = errors.New("migration action is not valid")
	ErrNoChange                  = migrate.ErrNoChange
	ErrNilVersion                = migrate.ErrNilVersion
	ErrMigrationPathNotExists    = errors.New("migration path do not exists")
)

type DBStatus struct {
	Version uint
	Dirty   bool
	Error   error
}

func checkPath(path string) (int, error) {
	if path == "" {
		return 0, ErrMigrationPathEmpty
	}

	dir, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, ErrMigrationPathNotExists
		}
	}

	count := len(dir)
	if count == 0 {
		return 0, ErrMigrationPathEmpty
	}

	return count, nil
}

func errLog(uri, path string, action Action, count int, err string) {
	fmt.Printf(
		"{\n\t**postgres_utils**\n\t[postgres_uri]: %s\n\t[migrate]: %s\n\t[action]: %s\n\t[count]: %d\n\t[error]: %s\n}\n",
		uri,
		path,
		action,
		count,
		err,
	)
}

func Migrate(uri string, path string, action Action) DBStatus {
	if path == "" {
		return DBStatus{Error: ErrMigrationPathNotExists}
	}

	instance, err := migrate.New(
		fmt.Sprintf("file://%s", path),
		uri,
	)
	if err != nil {
		return DBStatus{Error: err}
	}

	migrationsCount, err := checkPath(path)
	if err != nil {
		return DBStatus{Error: err}
	}

	switch action {
	case Up:
		if err := instance.Up(); err != nil {
			errLog(uri, fmt.Sprintf("file://%s", path), action, migrationsCount, err.Error())
			if errors.Is(err, migrate.ErrNoChange) {
				version, dirty, _ := instance.Version()
				return DBStatus{Version: version, Dirty: dirty, Error: migrate.ErrNoChange}
			}
			return DBStatus{Error: err}
		}
	case Down:
		if err := instance.Down(); err != nil {
			errLog(uri, fmt.Sprintf("file://%s", path), action, migrationsCount, err.Error())
			if errors.Is(err, migrate.ErrNoChange) {
				version, dirty, _ := instance.Version()
				return DBStatus{Version: version, Dirty: dirty, Error: migrate.ErrNoChange}
			}
			return DBStatus{Error: err}
		}
	default:
		return DBStatus{Error: ErrMigrationActionIsNotValid}
	}

	return DBStatus{}
}
