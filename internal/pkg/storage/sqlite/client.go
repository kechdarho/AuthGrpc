package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = createTables(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	go ClearTokens(ctx, db)
	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}
