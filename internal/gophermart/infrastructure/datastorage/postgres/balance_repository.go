package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type balanceRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) repository.BalanceRepository {
	return &balanceRepository{pool: pool}
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID int64) (*model.Balance, error) {
	query := `SELECT user_id, current, withdrawn FROM balances WHERE user_id = $1`
	balance := &model.Balance{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err != nil {
		balance = &model.Balance{
			UserID:    userID,
			Current:   0,
			Withdrawn: 0,
		}
	}
	return balance, nil
}

func (r *balanceRepository) Withdraw(ctx context.Context, userID int64, amount float64) error {
	query := `INSERT INTO balances (user_id, current, withdrawn) 
	          VALUES ($1, -$2, $2)
	          ON CONFLICT (user_id) 
	          DO UPDATE SET current = balances.current - $2, withdrawn = balances.withdrawn + $2`
	_, err := r.pool.Exec(ctx, query, userID, amount)
	return err
}

func (r *balanceRepository) Accrue(ctx context.Context, userID int64, amount float64) error {
	query := `INSERT INTO balances (user_id, current, withdrawn) 
	          VALUES ($1, $2, 0)
	          ON CONFLICT (user_id) 
	          DO UPDATE SET current = balances.current + $2`
	_, err := r.pool.Exec(ctx, query, userID, amount)
	return err
}


