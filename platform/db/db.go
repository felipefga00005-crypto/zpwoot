package db

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"zpwoot/internal/infra/db"
	"zpwoot/platform/logger"
)

type DB struct {
	*sqlx.DB
}

func New(databaseURL string) (*DB, error) {
	sqlxDB, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := sqlxDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: sqlxDB}, nil
}

// NewWithMigrations creates a new database connection and runs migrations
func NewWithMigrations(databaseURL string, logger *logger.Logger) (*DB, error) {
	database, err := New(databaseURL)
	if err != nil {
		return nil, err
	}

	// Run migrations automatically
	migrator := db.NewMigrator(database.DB.DB, logger)
	if err := migrator.RunMigrations(); err != nil {
		if closeErr := database.Close(); closeErr != nil {
			logger.Error("Failed to close database after migration error: " + closeErr.Error())
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but still panic with original error
				_ = rollbackErr // Explicitly ignore rollback error
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but keep original error
				_ = rollbackErr // Explicitly ignore rollback error
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Health checks database connectivity
func (db *DB) Health() error {
	return db.Ping()
}

// GetDB returns the underlying sqlx.DB instance
func (db *DB) GetDB() *sqlx.DB {
	return db.DB
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

// Get using this DB
func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Get(dest, query, args...)
}

// Select using this DB
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Select(dest, query, args...)
}
