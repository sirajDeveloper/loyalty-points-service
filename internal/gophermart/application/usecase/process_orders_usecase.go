package usecase

import (
	"context"
	"errors"
	"log"
	"time"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
)

type ProcessOrdersUseCase struct {
	outboxRepo    repository.OutboxRepository
	orderRepo     repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualService service.AccrualService
}

func NewProcessOrdersUseCase(
	outboxRepo repository.OutboxRepository,
	orderRepo repository.OrderRepository,
	balanceRepo repository.BalanceRepository,
	accrualService service.AccrualService,
) *ProcessOrdersUseCase {
	return &ProcessOrdersUseCase{
		outboxRepo:    outboxRepo,
		orderRepo:     orderRepo,
		balanceRepo:    balanceRepo,
		accrualService: accrualService,
	}
}

const (
	maxRetries = 3
	batchSize  = 10
)

func (uc *ProcessOrdersUseCase) ProcessPendingOrders(ctx context.Context) error {
	outboxes, err := uc.outboxRepo.FindPending(ctx, batchSize)
	if err != nil {
		return err
	}

	for _, outbox := range outboxes {
		if err := uc.processOrder(ctx, outbox); err != nil {
			log.Printf("failed to process order %d: %v", outbox.OrderID, err)
			if outbox.Retries >= maxRetries {
				if err := uc.outboxRepo.UpdateStatus(ctx, outbox.ID, model.OutboxStatusFailed); err != nil {
					log.Printf("failed to update outbox status to failed: %v", err)
				}
			} else {
				if err := uc.outboxRepo.IncrementRetries(ctx, outbox.ID); err != nil {
					log.Printf("failed to increment retries: %v", err)
				}
			}
			continue
		}

		if err := uc.outboxRepo.UpdateStatus(ctx, outbox.ID, model.OutboxStatusProcessed); err != nil {
			log.Printf("failed to update outbox status to processed: %v", err)
		}
	}

	return nil
}

func (uc *ProcessOrdersUseCase) processOrder(ctx context.Context, outbox *model.Outbox) error {
	order, err := uc.orderRepo.FindByID(ctx, outbox.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	accrualResp, err := uc.accrualService.GetOrderInfo(ctx, order.Number)
	if err != nil {
		if err.Error() == "rate limited" || err.Error() == "order not found in accrual system" {
			if err.Error() == "order not found in accrual system" {
				if updateErr := uc.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusInvalid, nil); updateErr != nil {
					return updateErr
				}
				return nil
			}
			return err
		}
		return err
	}

	var newStatus model.OrderStatus
	switch accrualResp.Status {
	case "REGISTERED":
		newStatus = model.OrderStatusNew
	case "PROCESSING":
		newStatus = model.OrderStatusProcessing
	case "INVALID":
		newStatus = model.OrderStatusInvalid
	case "PROCESSED":
		newStatus = model.OrderStatusProcessed
		if accrualResp.Accrual != nil {
			if err := uc.balanceRepo.Accrue(ctx, order.UserID, *accrualResp.Accrual); err != nil {
				return err
			}
		}
	default:
		return errors.New("unknown accrual status")
	}

	if err := uc.orderRepo.UpdateStatus(ctx, order.ID, newStatus, accrualResp.Accrual); err != nil {
		return err
	}

	return nil
}

func (uc *ProcessOrdersUseCase) StartWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := uc.ProcessPendingOrders(ctx); err != nil {
				log.Printf("error processing pending orders: %v", err)
			}
		}
	}
}

