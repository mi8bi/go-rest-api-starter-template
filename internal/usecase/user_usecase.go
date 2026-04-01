package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
	"github.com/mi8bi/go-rest-api-starter-template/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

// UserUsecase はユーザーのCRUDを担います。
type UserUsecase struct {
	users repository.UserRepository
}

func NewUserUsecase(users repository.UserRepository) *UserUsecase {
	return &UserUsecase{users: users}
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, err := uc.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (uc *UserUsecase) UpdateName(ctx context.Context, id int64, name string) (*domain.User, error) {
	u, err := uc.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	u.Name = name
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

func (uc *UserUsecase) Delete(ctx context.Context, id int64) error {
	u, err := uc.users.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if u == nil {
		return ErrUserNotFound
	}
	return uc.users.Delete(ctx, id)
}
