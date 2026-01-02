package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
)

type UploadOrderUseCase struct {
	unitOfWork     repository.UnitOfWork
	orderRepo      repository.OrderRepository
	outboxRepo     repository.OutboxRepository
	orderValidator service.OrderNumberValidator
}

func NewUploadOrderUseCase(
	unitOfWork repository.UnitOfWork,
	orderRepo repository.OrderRepository,
	outboxRepo repository.OutboxRepository,
	orderValidator service.OrderNumberValidator,
) *UploadOrderUseCase {
	return &UploadOrderUseCase{
		unitOfWork:     unitOfWork,
		orderRepo:      orderRepo,
		outboxRepo:     outboxRepo,
		orderValidator: orderValidator,
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
	if !uc.orderValidator.Validate(req.Number) {
		return nil, errors.New("invalid order number format")
	}

	existingOrder, err := uc.orderRepo.FindByNumber(ctx, req.Number)
	if err != nil {
		return nil, err
	}

	if existingOrder != nil {
		if existingOrder.CanBeUploadedBy(req.UserID) {
			return &UploadOrderResponse{Status: "already_uploaded"}, nil
		}
		return nil, errors.New("order number already exists for another user")
	}

	tx, err := uc.unitOfWork.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := model.NewOrder(req.UserID, req.Number)
	if err != nil {
		return nil, err
	}

	orderRepo := tx.OrderRepository()
	if err := orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	outbox := &model.Outbox{
		OrderID:   order.ID(),
		Status:    model.OutboxStatusPending,
		Retries:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	outboxRepo := tx.OutboxRepository()
	if err := outboxRepo.Create(ctx, outbox); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &UploadOrderResponse{Status: "accepted"}, nil
}
