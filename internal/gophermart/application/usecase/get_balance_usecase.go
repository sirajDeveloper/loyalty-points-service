package usecase

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type GetBalanceUseCase struct {
	balanceRepo repository.BalanceRepository
}

func NewGetBalanceUseCase(balanceRepo repository.BalanceRepository) *GetBalanceUseCase {
	return &GetBalanceUseCase{
		balanceRepo: balanceRepo,
	}
}

type GetBalanceRequest struct {
	UserID int64
}

type GetBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, req GetBalanceRequest) (*GetBalanceResponse, error) {
	balance, err := uc.balanceRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, nil
}


