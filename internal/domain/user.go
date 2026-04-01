package domain

import "time"

// User はアプリケーションのコアエンティティです。
// DB・HTTPに依存しない純粋なGoの型として定義します。
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
