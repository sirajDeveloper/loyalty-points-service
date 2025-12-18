package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type withdrawalRepository struct {
	pool *pgxpool.Pool
}

func NewWithdrawalRepository(pool *pgxpool.Pool) repository.WithdrawalRepository {
	return &withdrawalRepository{pool: pool}
}

func (r *withdrawalRepository) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.pool.QueryRow(ctx, query, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum, withdrawal.ProcessedAt).Scan(&withdrawal.ID)
	return err
}

func (r *withdrawalRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at 
	          FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		withdrawal := &model.Withdrawal{}
		err := rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, rows.Err()
}


