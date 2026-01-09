package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type balanceRepository struct {
	querier Querier
}

func NewBalanceRepository(pool *pgxpool.Pool) repository.BalanceRepository {
	return &balanceRepository{querier: pool}
}

func NewBalanceRepositoryTx(tx pgx.Tx) repository.BalanceRepository {
	return &balanceRepository{querier: tx}
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID int64) (*model.Balance, error) {
	query := `SELECT user_id, current, withdrawn FROM balances WHERE user_id = $1`
	var uid int64
	var current, withdrawn float64
	err := r.querier.QueryRow(ctx, query, userID).Scan(&uid, &current, &withdrawn)
	if err != nil {
		return model.NewBalance(userID), nil
	}
	return model.RestoreBalance(uid, current, withdrawn), nil
}

func (r *balanceRepository) Withdraw(ctx context.Context, userID int64, amount float64) error {
	query := `INSERT INTO balances (user_id, current, withdrawn) 
	          VALUES ($1, -$2::DECIMAL(10,2), $2::DECIMAL(10,2))
	          ON CONFLICT (user_id) 
	          DO UPDATE SET current = balances.current - $2::DECIMAL(10,2), withdrawn = balances.withdrawn + $2::DECIMAL(10,2)`
	_, err := r.querier.Exec(ctx, query, userID, amount)
	return err
}

func (r *balanceRepository) Accrue(ctx context.Context, userID int64, amount float64) error {
	query := `INSERT INTO balances (user_id, current, withdrawn) 
	          VALUES ($1, $2::DECIMAL(10,2), 0)
	          ON CONFLICT (user_id) 
	          DO UPDATE SET current = balances.current + $2::DECIMAL(10,2)`
	_, err := r.querier.Exec(ctx, query, userID, amount)
	return err
}
