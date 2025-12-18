package postgres

import (
	"context"
	"errors"
	"time"

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
	var id int64
	err := r.pool.QueryRow(ctx, query, order.UserID(), order.Number(), order.Status(), order.Accrual(), order.UploadedAt()).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("order number already exists")
		}
		return err
	}
	order.SetID(id)
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
		var id, uid int64
		var number string
		var status model.OrderStatus
		var accrual *float64
		var uploadedAt time.Time
		err := rows.Scan(&id, &uid, &number, &status, &accrual, &uploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, model.RestoreOrder(id, uid, number, status, accrual, uploadedAt))
	}

	return orders, rows.Err()
}

func (r *orderRepository) FindByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE number = $1`
	var id, userID int64
	var num string
	var status model.OrderStatus
	var accrual *float64
	var uploadedAt time.Time
	err := r.pool.QueryRow(ctx, query, number).Scan(&id, &userID, &num, &status, &accrual, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.RestoreOrder(id, userID, num, status, accrual, uploadedAt), nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE id = $1`
	var orderID, userID int64
	var number string
	var status model.OrderStatus
	var accrual *float64
	var uploadedAt time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(&orderID, &userID, &number, &status, &accrual, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.RestoreOrder(orderID, userID, number, status, accrual, uploadedAt), nil
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
		var id, userID int64
		var number string
		var status model.OrderStatus
		var accrual *float64
		var uploadedAt time.Time
		err := rows.Scan(&id, &userID, &number, &status, &accrual, &uploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, model.RestoreOrder(id, userID, number, status, accrual, uploadedAt))
	}

	return orders, rows.Err()
}

type orderRepositoryTx struct {
	tx pgx.Tx
}

func (r *orderRepositoryTx) Create(ctx context.Context, order *model.Order) error {
	query := `INSERT INTO orders (user_id, number, status, accrual, uploaded_at) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	err := r.tx.QueryRow(ctx, query, order.UserID(), order.Number(), order.Status(), order.Accrual(), order.UploadedAt()).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("order number already exists")
		}
		return err
	}
	order.SetID(id)
	return nil
}

func (r *orderRepositoryTx) FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := r.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var id, uid int64
		var number string
		var status model.OrderStatus
		var accrual *float64
		var uploadedAt time.Time
		err := rows.Scan(&id, &uid, &number, &status, &accrual, &uploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, model.RestoreOrder(id, uid, number, status, accrual, uploadedAt))
	}

	return orders, rows.Err()
}

func (r *orderRepositoryTx) FindByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE number = $1`
	var id, userID int64
	var num string
	var status model.OrderStatus
	var accrual *float64
	var uploadedAt time.Time
	err := r.tx.QueryRow(ctx, query, number).Scan(&id, &userID, &num, &status, &accrual, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.RestoreOrder(id, userID, num, status, accrual, uploadedAt), nil
}

func (r *orderRepositoryTx) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE id = $1`
	var orderID, userID int64
	var number string
	var status model.OrderStatus
	var accrual *float64
	var uploadedAt time.Time
	err := r.tx.QueryRow(ctx, query, id).Scan(&orderID, &userID, &number, &status, &accrual, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.RestoreOrder(orderID, userID, number, status, accrual, uploadedAt), nil
}

func (r *orderRepositoryTx) UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, status, accrual, orderID)
	return err
}

func (r *orderRepositoryTx) FindPending(ctx context.Context, limit int) ([]*model.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at 
	          FROM orders WHERE status IN ('NEW', 'PROCESSING') 
	          ORDER BY uploaded_at ASC LIMIT $1`
	rows, err := r.tx.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var id, userID int64
		var number string
		var status model.OrderStatus
		var accrual *float64
		var uploadedAt time.Time
		err := rows.Scan(&id, &userID, &number, &status, &accrual, &uploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, model.RestoreOrder(id, userID, number, status, accrual, uploadedAt))
	}

	return orders, rows.Err()
}
