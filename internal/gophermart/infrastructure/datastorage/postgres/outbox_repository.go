package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type outboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) repository.OutboxRepository {
	return &outboxRepository{pool: pool}
}

func (r *outboxRepository) Create(ctx context.Context, outbox *model.Outbox) error {
	query := `INSERT INTO outbox (order_id, status, retries, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.pool.QueryRow(ctx, query, outbox.OrderID, outbox.Status, outbox.Retries, outbox.CreatedAt, outbox.UpdatedAt).Scan(&outbox.ID)
	return err
}

func (r *outboxRepository) FindPending(ctx context.Context, limit int) ([]*model.Outbox, error) {
	query := `SELECT id, order_id, status, retries, created_at, updated_at 
	          FROM outbox WHERE status = 'PENDING' 
	          ORDER BY created_at ASC LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outboxes []*model.Outbox
	for rows.Next() {
		outbox := &model.Outbox{}
		err := rows.Scan(&outbox.ID, &outbox.OrderID, &outbox.Status, &outbox.Retries, &outbox.CreatedAt, &outbox.UpdatedAt)
		if err != nil {
			return nil, err
		}
		outboxes = append(outboxes, outbox)
	}

	return outboxes, rows.Err()
}

func (r *outboxRepository) UpdateStatus(ctx context.Context, outboxID int64, status model.OutboxStatus) error {
	query := `UPDATE outbox SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, outboxID)
	return err
}

func (r *outboxRepository) IncrementRetries(ctx context.Context, outboxID int64) error {
	query := `UPDATE outbox SET retries = retries + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, outboxID)
	return err
}


