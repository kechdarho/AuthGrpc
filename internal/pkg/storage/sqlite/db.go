package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

func createTables(ctx context.Context, db *sql.DB) error {
	queryUsers := `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login VARCHAR(40) NOT NULL,
			phone VARCHAR(30),
			email VARCHAR(255) NOT NULL,
			pass_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			UNIQUE(email),
			UNIQUE(login)
		)`

	_, err := db.ExecContext(ctx, queryUsers)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	queryResetTokens := `CREATE TABLE IF NOT EXISTS resetTokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`

	_, err = db.ExecContext(ctx, queryResetTokens)
	if err != nil {
		return fmt.Errorf("error creating resetTokens table: %v", err)
	}
	return nil
}

func ClearTokens(ctx context.Context, db *sql.DB) {
	for {
		time.Sleep(1 * time.Hour)
		_, err := db.ExecContext(ctx, "DELETE FROM resetTokens WHERE expires_at < DATETIME('now')")
		if err != nil {
			log.Printf("Failed to clean up expired tokens: %v", err)
		}
	}
}
