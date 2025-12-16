package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS player (
			key TEXT PRIMARY KEY,
			level INTEGER DEFAULT 1,
			xp_total INTEGER DEFAULT 0,
			xp_str INTEGER DEFAULT 0,
			xp_int INTEGER DEFAULT 0,
			xp_wis INTEGER DEFAULT 0,
			xp_art INTEGER DEFAULT 0,
			xp_home INTEGER DEFAULT 0,
			xp_out INTEGER DEFAULT 0,
			xp_read INTEGER DEFAULT 0,
			xp_cinema INTEGER DEFAULT 0,
			xp_career INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent_id INTEGER NULL,
			title TEXT NOT NULL,
			description TEXT,

			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			due_date DATETIME,

			difficulty INTEGER DEFAULT 1,
			attribute TEXT NOT NULL,
			attributes TEXT,
			xp_value INTEGER NOT NULL,

			is_project INTEGER DEFAULT 0,
			is_habit INTEGER DEFAULT 0,
			habit_interval TEXT,

			FOREIGN KEY(parent_id) REFERENCES tasks(id)
		);`,
		`CREATE TABLE IF NOT EXISTS blueprints (
			code TEXT PRIMARY KEY,
			status TEXT DEFAULT 'locked'
		);`,
		// Needed for habit decay (> 5 completions / 7 days) and auditing XP awarded.
		`CREATE TABLE IF NOT EXISTS task_completions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			completed_at DATETIME NOT NULL,
			difficulty INTEGER NOT NULL,
			xp_awarded INTEGER NOT NULL,
			FOREIGN KEY(task_id) REFERENCES tasks(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
		`CREATE INDEX IF NOT EXISTS idx_task_completions_task_id_completed_at ON task_completions(task_id, completed_at);`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	// Add new attribute columns to existing player tables (ignore if already exists)
	alterStmts := []string{
		`ALTER TABLE player ADD COLUMN xp_home INTEGER DEFAULT 0;`,
		`ALTER TABLE player ADD COLUMN xp_out INTEGER DEFAULT 0;`,
		`ALTER TABLE player ADD COLUMN xp_read INTEGER DEFAULT 0;`,
		`ALTER TABLE player ADD COLUMN xp_cinema INTEGER DEFAULT 0;`,
		`ALTER TABLE player ADD COLUMN xp_career INTEGER DEFAULT 0;`,
		// Multi-attribute support for tasks
		`ALTER TABLE tasks ADD COLUMN attributes TEXT;`,
		// Habit duration fields
		`ALTER TABLE tasks ADD COLUMN habit_start_date DATETIME;`,
		`ALTER TABLE tasks ADD COLUMN habit_end_date DATETIME;`,
		`ALTER TABLE tasks ADD COLUMN habit_goal INTEGER;`,
	}
	for _, stmt := range alterStmts {
		_, err := db.ExecContext(ctx, stmt)
		if err != nil && !strings.Contains(err.Error(), "duplicate column") {
			// Ignore "duplicate column" errors - column already exists
			return fmt.Errorf("migrate alter: %w", err)
		}
	}

	return nil
}
