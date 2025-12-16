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

func (r *CompletionRepo) CountSinceWithDifficulty(ctx context.Context, taskID int64, since time.Time, difficulty int) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM task_completions
		WHERE task_id = ? AND completed_at >= ? AND difficulty = ?
	`, taskID, since, difficulty)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, fmt.Errorf("completion count by difficulty: %w", err)
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
func (r *CompletionRepo) ListByTask(ctx context.Context, taskID int64) ([]TaskCompletion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, completed_at, difficulty, xp_awarded
		FROM task_completions
		WHERE task_id = ?
		ORDER BY completed_at ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("completion list: %w", err)
	}
	defer rows.Close()

	var out []TaskCompletion
	for rows.Next() {
		var tc TaskCompletion
		if err := rows.Scan(&tc.ID, &tc.TaskID, &tc.CompletedAt, &tc.Difficulty, &tc.XPAwarded); err != nil {
			return nil, fmt.Errorf("completion scan: %w", err)
		}
		out = append(out, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("completion rows: %w", err)
	}
	return out, nil
}

// ListInRange returns all completions between since and until (inclusive).
func (r *CompletionRepo) ListInRange(ctx context.Context, since, until time.Time) ([]TaskCompletion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, completed_at, difficulty, xp_awarded
		FROM task_completions
		WHERE completed_at >= ? AND completed_at <= ?
		ORDER BY completed_at ASC
	`, since, until)
	if err != nil {
		return nil, fmt.Errorf("completion list in range: %w", err)
	}
	defer rows.Close()

	var out []TaskCompletion
	for rows.Next() {
		var tc TaskCompletion
		if err := rows.Scan(&tc.ID, &tc.TaskID, &tc.CompletedAt, &tc.Difficulty, &tc.XPAwarded); err != nil {
			return nil, fmt.Errorf("completion scan: %w", err)
		}
		out = append(out, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("completion rows: %w", err)
	}
	return out, nil
}

// XPByDay returns a map of date (YYYY-MM-DD) to total XP earned that day.
func (r *CompletionRepo) XPByDay(ctx context.Context, since, until time.Time) (map[string]int, error) {
	completions, err := r.ListInRange(ctx, since, until)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int)
	for _, c := range completions {
		day := c.CompletedAt.Format("2006-01-02")
		result[day] += c.XPAwarded
	}
	return result, nil
}
