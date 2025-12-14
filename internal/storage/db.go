package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DefaultDBPath returns the default Questline DB location.
func DefaultDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(homeDir, ".questline.db"), nil
}

// OpenSQLite opens (and creates if missing) the SQLite database at the provided path.
// This is intentionally minimal during bootstrap; schema/migrations are implemented later.
func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}
package storage
package storage
































}	return db, nil	}		return nil, fmt.Errorf("ping sqlite: %w", err)		_ = db.Close()	if err := db.Ping(); err != nil {	}		return nil, fmt.Errorf("open sqlite: %w", err)	if err != nil {	db, err := sql.Open("sqlite", path)func OpenSQLite(path string) (*sql.DB, error) {// This is intentionally minimal during bootstrap; schema/migrations are implemented in the storage task.// OpenSQLite opens (and creates if missing) the SQLite database at the provided path.}	return filepath.Join(homeDir, ".questline.db"), nil	}		return "", fmt.Errorf("get home dir: %w", err)	if err != nil {	homeDir, err := os.UserHomeDir()func DefaultDBPath() (string, error) {// DefaultDBPath returns the default Questline DB location.)	_ "modernc.org/sqlite"	"path/filepath"	"os"	"fmt"	"database/sql"import (