package usecase

import (
	"context"
	"errors"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
)

type WithdrawUseCase struct {
	pool            *pgxpool.Pool
	balanceRepo     repository.BalanceRepository
	withdrawalRepo  repository.WithdrawalRepository
	luhnValidator   *service.LuhnValidator
}

func NewWithdrawUseCase(
	pool *pgxpool.Pool,
	balanceRepo repository.BalanceRepository,
	withdrawalRepo repository.WithdrawalRepository,
	luhnValidator *service.LuhnValidator,
) *WithdrawUseCase {
	return &WithdrawUseCase{
		pool:           pool,
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		luhnValidator:  luhnValidator,
	}
}

type WithdrawRequest struct {
	UserID     int64
	Order      string
	Sum        float64
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

	if balance.Current < req.Sum {
		return nil, errors.New("insufficient funds")
	}

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	withdrawal := &model.Withdrawal{
		UserID:      req.UserID,
		OrderNumber: req.Order,
		Sum:         req.Sum,
		ProcessedAt: time.Now(),
	}

	withdrawalQuery := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
	                    VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRow(ctx, withdrawalQuery, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum, withdrawal.ProcessedAt).Scan(&withdrawal.ID)
	if err != nil {
		return nil, err
	}

	balanceQuery := `INSERT INTO balances (user_id, current, withdrawn) 
	                 VALUES ($1, -$2, $2)
	                 ON CONFLICT (user_id) 
	                 DO UPDATE SET current = balances.current - $2, withdrawn = balances.withdrawn + $2`
	_, err = tx.Exec(ctx, balanceQuery, req.UserID, req.Sum)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &WithdrawResponse{Success: true}, nil
}


