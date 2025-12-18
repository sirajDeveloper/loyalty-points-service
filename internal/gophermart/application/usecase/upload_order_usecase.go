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

type UploadOrderUseCase struct {
	pool         *pgxpool.Pool
	orderRepo    repository.OrderRepository
	outboxRepo   repository.OutboxRepository
	luhnValidator *service.LuhnValidator
}

func NewUploadOrderUseCase(
	pool *pgxpool.Pool,
	orderRepo repository.OrderRepository,
	outboxRepo repository.OutboxRepository,
	luhnValidator *service.LuhnValidator,
) *UploadOrderUseCase {
	return &UploadOrderUseCase{
		pool:         pool,
		orderRepo:    orderRepo,
		outboxRepo:   outboxRepo,
		luhnValidator: luhnValidator,
	}
}

type UploadOrderRequest struct {
	UserID int64
	Number string
}

type UploadOrderResponse struct {
	Status string
}

func (uc *UploadOrderUseCase) Execute(ctx context.Context, req UploadOrderRequest) (*UploadOrderResponse, error) {
	if !uc.luhnValidator.Validate(req.Number) {
		return nil, errors.New("invalid order number format")
	}

	existingOrder, err := uc.orderRepo.FindByNumber(ctx, req.Number)
	if err != nil {
		return nil, err
	}

	if existingOrder != nil {
		if existingOrder.UserID == req.UserID {
			return &UploadOrderResponse{Status: "already_uploaded"}, nil
		}
		return nil, errors.New("order number already exists for another user")
	}

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order := &model.Order{
		UserID:     req.UserID,
		Number:     req.Number,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	orderQuery := `INSERT INTO orders (user_id, number, status, accrual, uploaded_at) 
	               VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRow(ctx, orderQuery, order.UserID, order.Number, order.Status, order.Accrual, order.UploadedAt).Scan(&order.ID)
	if err != nil {
		return nil, err
	}

	outbox := &model.Outbox{
		OrderID:   order.ID,
		Status:    model.OutboxStatusPending,
		Retries:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	outboxQuery := `INSERT INTO outbox (order_id, status, retries, created_at, updated_at) 
	                VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRow(ctx, outboxQuery, outbox.OrderID, outbox.Status, outbox.Retries, outbox.CreatedAt, outbox.UpdatedAt).Scan(&outbox.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &UploadOrderResponse{Status: "accepted"}, nil
}

