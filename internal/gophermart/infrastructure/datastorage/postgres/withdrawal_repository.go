package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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
	var id int64
	err := r.pool.QueryRow(ctx, query, withdrawal.UserID(), withdrawal.OrderNumber(), withdrawal.Sum(), withdrawal.ProcessedAt()).Scan(&id)
	if err != nil {
		return err
	}
	withdrawal.SetID(id)
	return nil
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
		var id, uid int64
		var orderNumber string
		var sum float64
		var processedAt time.Time
		err := rows.Scan(&id, &uid, &orderNumber, &sum, &processedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, model.RestoreWithdrawal(id, uid, orderNumber, sum, processedAt))
	}

	return withdrawals, rows.Err()
}

type withdrawalRepositoryTx struct {
	tx pgx.Tx
}

func (r *withdrawalRepositoryTx) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	var id int64
	err := r.tx.QueryRow(ctx, query, withdrawal.UserID(), withdrawal.OrderNumber(), withdrawal.Sum(), withdrawal.ProcessedAt()).Scan(&id)
	if err != nil {
		return err
	}
	withdrawal.SetID(id)
	return nil
}

func (r *withdrawalRepositoryTx) FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at 
	          FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := r.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var id, uid int64
		var orderNumber string
		var sum float64
		var processedAt time.Time
		err := rows.Scan(&id, &uid, &orderNumber, &sum, &processedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, model.RestoreWithdrawal(id, uid, orderNumber, sum, processedAt))
	}

	return withdrawals, rows.Err()
}
