package postgres

import (
	"errors"
	"executor/internal/core/config"
	pu "executor/pkg/postgres_utils"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrConnect = errors.New("could not connect to database")
	ErrPing    = errors.New("could not ping database")
)

type PostgresRepository struct {
	db             *sqlx.DB
	uri            string
	migrationsPath string
	version        uint
	dirty          bool
}

func NewPostgresRepository(cfg *config.ExecutorConfig) *PostgresRepository {
	repo := &PostgresRepository{}
	if err := repo.InvokeConnect(cfg); err != nil {
		panic(err)
	}
	return repo
}

func (repo *PostgresRepository) InvokeConnect(cfg *config.ExecutorConfig) error {
	postgres_uri := pu.ParseURI(
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
		cfg.Postgres.Port,
	)
	repo.uri = postgres_uri
	repo.migrationsPath = cfg.Postgres.MigrationsPath

	db, err := sqlx.Open("postgres", postgres_uri)
	if err != nil {
		return ErrConnect
	}
	repo.db = db
	if err := repo.PingTest(); err != nil {
		panic(err)
	}
	return nil
}

func (repo *PostgresRepository) PingTest() error {
	max_errs := 5
	errs := 0
	timeout := 1 * time.Second
	for max_errs > 0 {
		if err := repo.db.Ping(); err != nil {
			fmt.Printf("could not ping database: %s\n", err.Error())
			fmt.Printf("retrying in %s\n", timeout)
			max_errs--
			errs++
			time.Sleep(timeout)
		}
		max_errs = 0
		errs = 0
	}
	if errs == 0 {
		return nil
	}
	return fmt.Errorf("%w: postgres_uri: %s", ErrPing, repo.uri)
}

func (repo *PostgresRepository) Close() {
	repo.db.Close()
}

func (repo *PostgresRepository) Migrate() error {
	if status := pu.Migrate(repo.uri, repo.migrationsPath, pu.Up); status.Error != nil {
		if status.Error != pu.ErrNoChange {
			return status.Error
		}
	} else {
		repo.version = status.Version
		repo.dirty = status.Dirty
	}
	return nil
}
