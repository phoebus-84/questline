package root

import (
	"context"
	"database/sql"

	"questline/internal/engine"
	"questline/internal/storage"
)

func openDB(ctx context.Context) (*sql.DB, func(), error) {
	path, err := storage.ResolveDBPath()
	if err != nil {
		return nil, nil, err
	}
	db, err := storage.Open(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = db.Close()
	}
	return db, cleanup, nil
}

func openService(ctx context.Context) (*engine.Service, func(), error) {
	db, cleanup, err := openDB(ctx)
	if err != nil {
		return nil, nil, err
	}
	return engine.NewService(db), cleanup, nil
}
