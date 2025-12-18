package usecase

import (
	"context"
	"errors"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
)

type WithdrawUseCase struct {
	unitOfWork     repository.UnitOfWork
	balanceRepo    repository.BalanceRepository
	withdrawalRepo repository.WithdrawalRepository
	luhnValidator  *service.LuhnValidator
}

func NewWithdrawUseCase(
	unitOfWork repository.UnitOfWork,
	balanceRepo repository.BalanceRepository,
	withdrawalRepo repository.WithdrawalRepository,
	luhnValidator *service.LuhnValidator,
) *WithdrawUseCase {
	return &WithdrawUseCase{
		unitOfWork:     unitOfWork,
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		luhnValidator:  luhnValidator,
	}
}

type WithdrawRequest struct {
	UserID int64
	Order  string
	Sum    float64
}

type WithdrawResponse struct {
	Success bool
}

func (uc *WithdrawUseCase) Execute(ctx context.Context, req WithdrawRequest) (*WithdrawResponse, error) {
	if !uc.luhnValidator.Validate(req.Order) {
		return nil, errors.New("invalid order number format")
	}

	balance, err := uc.balanceRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if !balance.CanWithdraw(req.Sum) {
		return nil, errors.New("insufficient funds")
	}

	tx, err := uc.unitOfWork.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	withdrawal, err := model.NewWithdrawal(req.UserID, req.Order, req.Sum)
	if err != nil {
		return nil, err
	}

	withdrawalRepo := tx.WithdrawalRepository()
	if err := withdrawalRepo.Create(ctx, withdrawal); err != nil {
		return nil, err
	}

	balanceRepo := tx.BalanceRepository()
	if err := balanceRepo.Withdraw(ctx, req.UserID, req.Sum); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &WithdrawResponse{Success: true}, nil
}
