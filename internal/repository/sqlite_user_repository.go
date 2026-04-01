package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
)

// SQLiteUserRepository はUserRepositoryのSQLite実装です。
type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = ?`, email)
	return scanUser(row)
}

func (r *SQLiteUserRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		user.Name, user.Email, user.PasswordHash, now, now,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *domain.User) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = ?, email = ?, password_hash = ?, updated_at = ? WHERE id = ?`,
		user.Name, user.Email, user.PasswordHash, now, user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	user.UpdatedAt = now
	return nil
}

func (r *SQLiteUserRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

// scanUser は *sql.Row をdomain.Userにマップします。
func scanUser(row *sql.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // not found → nil を返す（usecase側でハンドリング）
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
