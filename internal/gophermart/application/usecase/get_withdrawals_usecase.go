package usecase

import (
	"context"
	"time"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type GetWithdrawalsUseCase struct {
	withdrawalRepo repository.WithdrawalRepository
}

func NewGetWithdrawalsUseCase(withdrawalRepo repository.WithdrawalRepository) *GetWithdrawalsUseCase {
	return &GetWithdrawalsUseCase{
		withdrawalRepo: withdrawalRepo,
	}
}

type GetWithdrawalsRequest struct {
	UserID int64
}

func (uc *GetWithdrawalsUseCase) Execute(ctx context.Context, req GetWithdrawalsRequest) ([]*WithdrawalResponse, error) {
	withdrawals, err := uc.withdrawalRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return []*WithdrawalResponse{}, nil
	}

	response := make([]*WithdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		response = append(response, &WithdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	return response, nil
}

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

