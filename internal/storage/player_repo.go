package storage

import (
	"context"
	"database/sql"
	"fmt"
)

const MainPlayerKey = "main_user"

type PlayerRepo struct {
	db *sql.DB
}

func NewPlayerRepo(db *sql.DB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

func (r *PlayerRepo) Get(ctx context.Context, key string) (*Player, error) {
	row := r.db.QueryRowContext(ctx, `SELECT key, level, xp_total, xp_str, xp_int, xp_wis, xp_art FROM player WHERE key = ?`, key)

	var p Player
	if err := row.Scan(&p.Key, &p.Level, &p.XPTotal, &p.XPStr, &p.XPInt, &p.XPWis, &p.XPArt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("player get: %w", err)
	}
	return &p, nil
}

func (r *PlayerRepo) GetOrCreateMain(ctx context.Context) (*Player, error) {
	p, err := r.Get(ctx, MainPlayerKey)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}

	if _, err := r.db.ExecContext(ctx, `INSERT INTO player (key) VALUES (?)`, MainPlayerKey); err != nil {
		return nil, fmt.Errorf("player insert: %w", err)
	}
	return r.Get(ctx, MainPlayerKey)
}

func (r *PlayerRepo) Update(ctx context.Context, p *Player) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE player
		SET level = ?, xp_total = ?, xp_str = ?, xp_int = ?, xp_wis = ?, xp_art = ?
		WHERE key = ?
	`, p.Level, p.XPTotal, p.XPStr, p.XPInt, p.XPWis, p.XPArt, p.Key)
	if err != nil {
		return fmt.Errorf("player update: %w", err)
	}
	return nil
}
