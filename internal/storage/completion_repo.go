package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CompletionRepo struct {
	db *sql.DB
}

func NewCompletionRepo(db *sql.DB) *CompletionRepo {
	return &CompletionRepo{db: db}
}

func (r *CompletionRepo) Insert(ctx context.Context, taskID int64, completedAt time.Time, difficulty int, xpAwarded int) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO task_completions (task_id, completed_at, difficulty, xp_awarded)
		VALUES (?, ?, ?, ?)
	`, taskID, completedAt, difficulty, xpAwarded)
	if err != nil {
		return 0, fmt.Errorf("completion insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("completion last insert id: %w", err)
	}
	return id, nil
}

func (r *CompletionRepo) CountSince(ctx context.Context, taskID int64, since time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM task_completions
		WHERE task_id = ? AND completed_at >= ?
	`, taskID, since)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, fmt.Errorf("completion count: %w", err)
	}
	return n, nil
}

func (r *CompletionRepo) Last(ctx context.Context, taskID int64) (*TaskCompletion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, task_id, completed_at, difficulty, xp_awarded
		FROM task_completions
		WHERE task_id = ?
		ORDER BY completed_at DESC
		LIMIT 1
	`, taskID)
	var tc TaskCompletion
	if err := row.Scan(&tc.ID, &tc.TaskID, &tc.CompletedAt, &tc.Difficulty, &tc.XPAwarded); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("completion last: %w", err)
	}
	return &tc, nil
}
