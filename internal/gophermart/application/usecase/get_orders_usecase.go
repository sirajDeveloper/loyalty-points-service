package usecase

import (
	"context"
	"time"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type GetOrdersUseCase struct {
	orderRepo repository.OrderRepository
}

func NewGetOrdersUseCase(orderRepo repository.OrderRepository) *GetOrdersUseCase {
	return &GetOrdersUseCase{
		orderRepo: orderRepo,
	}
}

type GetOrdersRequest struct {
	UserID int64
}

func (uc *GetOrdersUseCase) Execute(ctx context.Context, req GetOrdersRequest) ([]*OrderResponse, error) {
	orders, err := uc.orderRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return []*OrderResponse{}, nil
	}

	response := make([]*OrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, &OrderResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	return response, nil
}

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

