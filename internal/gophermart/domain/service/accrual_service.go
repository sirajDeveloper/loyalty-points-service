package service

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

type AccrualService interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*model.AccrualResponse, error)
}

