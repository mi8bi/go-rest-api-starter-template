package infra

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB はSQLiteのDBコネクションを生成し、テーブルを初期化します。
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			name          TEXT    NOT NULL,
			email         TEXT    NOT NULL UNIQUE,
			password_hash TEXT    NOT NULL,
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT    PRIMARY KEY,
			user_id    INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	return err
}
