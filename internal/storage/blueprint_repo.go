package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type BlueprintRepo struct {
	db *sql.DB
}

func NewBlueprintRepo(db *sql.DB) *BlueprintRepo {
	return &BlueprintRepo{db: db}
}

func (r *BlueprintRepo) Get(ctx context.Context, code string) (*Blueprint, error) {
	row := r.db.QueryRowContext(ctx, `SELECT code, status FROM blueprints WHERE code = ?`, code)
	var b Blueprint
	if err := row.Scan(&b.Code, &b.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("blueprint get: %w", err)
	}
	return &b, nil
}

func (r *BlueprintRepo) Upsert(ctx context.Context, b Blueprint) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO blueprints (code, status) VALUES (?, ?)
		ON CONFLICT(code) DO UPDATE SET status = excluded.status
	`, b.Code, b.Status)
	if err != nil {
		return fmt.Errorf("blueprint upsert: %w", err)
	}
	return nil
}

func (r *BlueprintRepo) ListByStatus(ctx context.Context, status string) ([]Blueprint, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code, status FROM blueprints WHERE status = ? ORDER BY code ASC`, status)
	if err != nil {
		return nil, fmt.Errorf("blueprint list: %w", err)
	}
	defer rows.Close()

	var out []Blueprint
	for rows.Next() {
		var b Blueprint
		if err := rows.Scan(&b.Code, &b.Status); err != nil {
			return nil, fmt.Errorf("blueprint scan: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("blueprint rows: %w", err)
	}
	return out, nil
}
