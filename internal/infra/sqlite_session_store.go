package infra

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const sessionTTL = 24 * time.Hour

// SQLiteSessionStore はセッションをSQLiteに保存します。
type SQLiteSessionStore struct {
	db *sql.DB
}

func NewSQLiteSessionStore(db *sql.DB) *SQLiteSessionStore {
	return &SQLiteSessionStore{db: db}
}

// Create は新しいセッショントークンを生成してDBに保存し、トークンを返します。
func (s *SQLiteSessionStore) Create(ctx context.Context, userID int64) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	expiresAt := time.Now().Add(sessionTTL)
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

// GetUserID はトークンが有効なら対応するuserIDを返します。
func (s *SQLiteSessionStore) GetUserID(ctx context.Context, token string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE token = ?`, token,
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil // not found
	}
	if err != nil {
		return 0, fmt.Errorf("query session: %w", err)
	}
	if time.Now().After(expiresAt) {
		_ = s.Delete(ctx, token) // 期限切れは削除
		return 0, nil
	}
	return userID, nil
}

// Delete はセッションを削除します（ログアウト用）。
func (s *SQLiteSessionStore) Delete(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
