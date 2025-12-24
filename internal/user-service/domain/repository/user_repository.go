package repository

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	ExistsByLogin(ctx context.Context, login string) (bool, error)
}


