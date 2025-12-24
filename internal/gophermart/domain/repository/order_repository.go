package repository

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	FindByNumber(ctx context.Context, number string) (*model.Order, error)
	FindByID(ctx context.Context, id int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus, accrual *float64) error
	FindPending(ctx context.Context, limit int) ([]*model.Order, error)
}

