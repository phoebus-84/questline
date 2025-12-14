package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	// EnvDBPath overrides the default DB location.
	EnvDBPath = "QL_DB_PATH"
)

// DefaultDBPath returns the default Questline DB location.
func DefaultDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(homeDir, ".questline.db"), nil
}

// ResolveDBPath returns the database path, preferring EnvDBPath when set.
func ResolveDBPath() (string, error) {
	if v := os.Getenv(EnvDBPath); v != "" {
		return v, nil
	}
	return DefaultDBPath()
}

// Open opens (and creates if missing) the SQLite database and runs migrations.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign_keys pragma: %w", err)
	}
	if err := Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}