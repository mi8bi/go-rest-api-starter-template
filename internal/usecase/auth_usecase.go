package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
	"github.com/mi8bi/go-rest-api-starter-template/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// SessionStore はセッション操作の抽象（interfaceはここで定義）。
type SessionStore interface {
	Create(ctx context.Context, userID int64) (string, error)
	GetUserID(ctx context.Context, token string) (int64, error)
	Delete(ctx context.Context, token string) error
}

// AuthUsecase はログイン・登録・ログアウトを担います。
type AuthUsecase struct {
	users    repository.UserRepository
	sessions SessionStore
}

func NewAuthUsecase(users repository.UserRepository, sessions SessionStore) *AuthUsecase {
	return &AuthUsecase{users: users, sessions: sessions}
}

// Register は新規ユーザーを登録します。
func (uc *AuthUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	existing, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find by email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login はメール・パスワードを検証してセッショントークンを返します。
func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (token string, err error) {
	u, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err = uc.sessions.Create(ctx, u.ID)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

// Logout はセッションを削除します。
func (uc *AuthUsecase) Logout(ctx context.Context, token string) error {
	return uc.sessions.Delete(ctx, token)
}

// Me はセッショントークンからユーザーを返します。
func (uc *AuthUsecase) Me(ctx context.Context, token string) (*domain.User, error) {
	userID, err := uc.sessions.GetUserID(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if userID == 0 {
		return nil, nil // セッション無効
	}
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}
