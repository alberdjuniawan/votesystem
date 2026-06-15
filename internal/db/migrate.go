package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

type Migration struct {
	Version string
	Name    string
	SQL     string
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	if len(migrations) == 0 {
		return nil
	}

	if err := ensureSchemaTable(ctx, pool); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", m.Version, err)
		}

		if _, err := tx.Exec(ctx, m.SQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migration %s (%s) failed: %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)`,
			m.Version, m.Name, time.Now().UTC(),
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.Version, err)
		}
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		data, err := fs.ReadFile(migrationFiles, "migrations/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		version := parts[0]
		name := strings.TrimSuffix(entry.Name(), ".up.sql")
		if len(parts) > 1 {
			name = strings.TrimSuffix(parts[1], ".up.sql")
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(data),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func ensureSchemaTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(20)  PRIMARY KEY,
		name       TEXT         NOT NULL,
		applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	)`)
	return err
}

func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}
