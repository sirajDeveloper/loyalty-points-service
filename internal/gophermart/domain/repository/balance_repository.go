package repository

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*model.Balance, error)
	Withdraw(ctx context.Context, userID int64, amount float64) error
	Accrue(ctx context.Context, userID int64, amount float64) error
}


