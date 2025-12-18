package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type orderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) repository.OrderRepository {
	return &orderRepository{pool: pool}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `INSERT INTO orders (user_id, number, status, accrual, uploaded_at) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.pool.QueryRow(ctx, query, order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt).Scan(&order.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("order number already exists")
		}
		return err
	}
	return nil
}

func (r *orderRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (r *orderRepository) FindByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE number = $1`
	order := &model.Order{}
	err := r.pool.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE id = $1`
	order := &model.Order{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, accrual, orderID)
	return err
}

func (r *orderRepository) FindPending(ctx context.Context, limit int) ([]*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE status IN ('NEW', 'PROCESSING') 
	          ORDER BY uploaded_at ASC LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

