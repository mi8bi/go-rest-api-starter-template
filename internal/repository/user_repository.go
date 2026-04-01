package repository

import (
	"context"

	"github.com/mi8bi/go-rest-api-starter-template/internal/domain"
)

// UserRepository はユーザーの永続化操作を抽象化します。
// interfaceはこれを使うusecase層に近い場所（repositoryパッケージ）に定義します。
// 実装はinfra層に置きます。
type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}
