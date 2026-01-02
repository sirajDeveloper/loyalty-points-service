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
	querier Querier
}

func NewWithdrawalRepository(pool *pgxpool.Pool) repository.WithdrawalRepository {
	return &withdrawalRepository{querier: pool}
}

func NewWithdrawalRepositoryTx(tx pgx.Tx) repository.WithdrawalRepository {
	return &withdrawalRepository{querier: tx}
}

func (r *withdrawalRepository) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	var id int64
	err := r.querier.QueryRow(ctx, query, withdrawal.UserID(), withdrawal.OrderNumber(), withdrawal.Sum(), withdrawal.ProcessedAt()).Scan(&id)
	if err != nil {
		return err
	}
	withdrawal.SetID(id)
	return nil
}

func (r *withdrawalRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at 
	          FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	return scanRows(rows, func(rows pgx.Rows) (*model.Withdrawal, error) {
		var id, uid int64
		var orderNumber string
		var sum float64
		var processedAt time.Time
		err := rows.Scan(&id, &uid, &orderNumber, &sum, &processedAt)
		if err != nil {
			return nil, err
		}
		return model.RestoreWithdrawal(id, uid, orderNumber, sum, processedAt), nil
	})
}

func (r *withdrawalRepository) FindByUserIDIterator(ctx context.Context, userID int64) (Iterator[*model.Withdrawal], error) {
	query := `SELECT id, user_id, order_number, sum, processed_at 
	          FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	return NewIterator(rows, func(rows pgx.Rows) (*model.Withdrawal, error) {
		var id, uid int64
		var orderNumber string
		var sum float64
		var processedAt time.Time
		err := rows.Scan(&id, &uid, &orderNumber, &sum, &processedAt)
		if err != nil {
			return nil, err
		}
		return model.RestoreWithdrawal(id, uid, orderNumber, sum, processedAt), nil
	}), nil
}
