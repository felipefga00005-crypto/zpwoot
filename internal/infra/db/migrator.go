package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *time.Time
}

type Migrator struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewMigrator(db *sql.DB, logger *logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

func (m *Migrator) RunMigrations() error {
	m.logger.Info("Starting database migrations...")

	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	pendingCount := 0
	for _, migration := range migrations {
		if !m.isMigrationApplied(migration.Version, appliedMigrations) {
			if err := m.executeMigration(migration); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}
			pendingCount++
		}
	}

	if pendingCount > 0 {
		m.logger.InfoWithFields("Database migrations completed", map[string]interface{}{
			"migrations_applied": pendingCount,
			"total_migrations":   len(migrations),
		})
	} else {
		m.logger.Info("Database is up to date, no migrations needed")
	}

	return nil
}

func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS "zpMigrations" (
			"version" INTEGER PRIMARY KEY,
			"name" VARCHAR(255) NOT NULL,
			"appliedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS "idx_zp_migrations_applied_at" ON "zpMigrations" ("appliedAt");
		
		COMMENT ON TABLE "zpMigrations" IS 'Database migrations tracking table';
		COMMENT ON COLUMN "zpMigrations"."version" IS 'Migration version number';
		COMMENT ON COLUMN "zpMigrations"."name" IS 'Migration name';
		COMMENT ON COLUMN "zpMigrations"."appliedAt" IS 'When migration was applied';
	`

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func (m *Migrator) loadMigrations() ([]*Migration, error) {
	var migrations []*Migration

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationFiles := make(map[int]map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			m.logger.WarnWithFields("Skipping invalid migration file", map[string]interface{}{
				"filename": entry.Name(),
				"error":    err.Error(),
			})
			continue
		}

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		if migrationFiles[version] == nil {
			migrationFiles[version] = make(map[string]string)
		}

		if strings.Contains(entry.Name(), ".up.sql") {
			migrationFiles[version]["up"] = string(content)
			nameParts := strings.Split(entry.Name(), "_")
			if len(nameParts) > 1 {
				name := strings.Join(nameParts[1:], "_")
				name = strings.TrimSuffix(name, ".up.sql")
				migrationFiles[version]["name"] = name
			}
		} else if strings.Contains(entry.Name(), ".down.sql") {
			migrationFiles[version]["down"] = string(content)
		}
	}

	for version, files := range migrationFiles {
		migration := &Migration{
			Version: version,
			Name:    files["name"],
			UpSQL:   files["up"],
			DownSQL: files["down"],
		}

		if migration.UpSQL == "" {
			m.logger.WarnWithFields("Migration missing up.sql file", map[string]interface{}{
				"version": version,
			})
			continue
		}

		migrations = append(migrations, migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) getAppliedMigrations() (map[int]bool, error) {
	query := `SELECT "version" FROM "zpMigrations" ORDER BY "version"`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.logger.Error("Failed to close rows: " + err.Error())
		}
	}()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

func (m *Migrator) isMigrationApplied(version int, appliedMigrations map[int]bool) bool {
	return appliedMigrations[version]
}

func (m *Migrator) executeMigration(migration *Migration) error {
	m.logger.InfoWithFields("Applying migration", map[string]interface{}{
		"version": migration.Version,
		"name":    migration.Name,
	})

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				m.logger.Error("Failed to rollback transaction: " + rollbackErr.Error())
			}
		}
	}()

	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	insertQuery := `
		INSERT INTO "zpMigrations" ("version", "name", "appliedAt")
		VALUES ($1, $2, NOW())
	`
	if _, err := tx.Exec(insertQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}
	committed = true

	m.logger.InfoWithFields("Migration applied successfully", map[string]interface{}{
		"version": migration.Version,
		"name":    migration.Name,
	})

	return nil
}

func (m *Migrator) Rollback() error {
	m.logger.Info("Rolling back last migration...")

	query := `
		SELECT "version", "name" 
		FROM "zpMigrations" 
		ORDER BY "version" DESC 
		LIMIT 1
	`

	var version int
	var name string
	err := m.db.QueryRow(query).Scan(&version, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.Version == version {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found in files", version)
	}

	if targetMigration.DownSQL == "" {
		return fmt.Errorf("migration %d has no down SQL", version)
	}

	m.logger.InfoWithFields("Rolling back migration", map[string]interface{}{
		"version": version,
		"name":    name,
	})

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				m.logger.Error("Failed to rollback transaction: " + rollbackErr.Error())
			}
		}
	}()

	if _, err := tx.Exec(targetMigration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	deleteQuery := `DELETE FROM "zpMigrations" WHERE "version" = $1`
	if _, err := tx.Exec(deleteQuery, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}
	committed = true

	m.logger.InfoWithFields("Migration rolled back successfully", map[string]interface{}{
		"version": version,
		"name":    name,
	})

	return nil
}

func (m *Migrator) GetMigrationStatus() ([]*Migration, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if appliedMigrations[migration.Version] {
			now := time.Now()
			migration.AppliedAt = &now
		}
	}

	return migrations, nil
}
