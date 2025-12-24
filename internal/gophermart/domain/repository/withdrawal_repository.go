package repository

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *model.Withdrawal) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}


