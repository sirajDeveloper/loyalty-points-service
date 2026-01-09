package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
)

type ProcessOrdersUseCase struct {
	outboxRepo     repository.OutboxRepository
	orderRepo      repository.OrderRepository
	balanceRepo    repository.BalanceRepository
	accrualService service.AccrualService
}

func NewProcessOrdersUseCase(
	outboxRepo repository.OutboxRepository,
	orderRepo repository.OrderRepository,
	balanceRepo repository.BalanceRepository,
	accrualService service.AccrualService,
) *ProcessOrdersUseCase {
	return &ProcessOrdersUseCase{
		outboxRepo:     outboxRepo,
		orderRepo:      orderRepo,
		balanceRepo:    balanceRepo,
		accrualService: accrualService,
	}
}

const (
	maxRetries     = 3
	batchSize      = 10
	maxConcurrency = 5
)

func (uc *ProcessOrdersUseCase) ProcessPendingOrders(ctx context.Context) error {
	outboxes, err := uc.outboxRepo.FindPending(ctx, batchSize)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pending outboxes", "error", err)
		return err
	}

	if len(outboxes) == 0 {
		return nil
	}

	slog.InfoContext(ctx, "processing pending orders", "count", len(outboxes))

	g, gCtx := errgroup.WithContext(ctx)

	sem := make(chan struct{}, maxConcurrency)

	for _, outbox := range outboxes {
		outbox := outbox

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := uc.processOrder(gCtx, outbox); err != nil {
				slog.WarnContext(gCtx, "failed to process order",
					"order_id", outbox.OrderID,
					"outbox_id", outbox.ID,
					"retries", outbox.Retries,
					"error", err,
				)

				if outbox.Retries >= maxRetries {
					if updateErr := uc.outboxRepo.UpdateStatus(gCtx, outbox.ID, model.OutboxStatusFailed); updateErr != nil {
						slog.ErrorContext(gCtx, "failed to update outbox status to failed",
							"outbox_id", outbox.ID,
							"error", updateErr,
						)
					} else {
						slog.InfoContext(gCtx, "outbox marked as failed",
							"outbox_id", outbox.ID,
							"retries", outbox.Retries,
						)
					}
				} else {
					if updateErr := uc.outboxRepo.IncrementRetries(gCtx, outbox.ID); updateErr != nil {
						slog.ErrorContext(gCtx, "failed to increment retries",
							"outbox_id", outbox.ID,
							"error", updateErr,
						)
					}
				}
				return nil
			}

			if err := uc.outboxRepo.UpdateStatus(gCtx, outbox.ID, model.OutboxStatusProcessed); err != nil {
				slog.ErrorContext(gCtx, "failed to update outbox status to processed",
					"outbox_id", outbox.ID,
					"error", err,
				)
			} else {
				slog.InfoContext(gCtx, "order processed successfully",
					"order_id", outbox.OrderID,
					"outbox_id", outbox.ID,
				)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "error in order processing group", "error", err)
		return err
	}

	slog.InfoContext(ctx, "finished processing pending orders", "count", len(outboxes))
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

	accrualResp, err := uc.accrualService.GetOrderInfo(ctx, order.Number())
	if err != nil {
		if err.Error() == "rate limited" || err.Error() == "order not found in accrual system" {
			if err.Error() == "order not found in accrual system" {
				if updateErr := order.UpdateStatus(model.OrderStatusInvalid, nil); updateErr != nil {
					return updateErr
				}
				if updateErr := uc.orderRepo.UpdateStatus(ctx, order.ID(), model.OrderStatusInvalid, nil); updateErr != nil {
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
			if err := uc.balanceRepo.Accrue(ctx, order.UserID(), *accrualResp.Accrual); err != nil {
				return err
			}
		}
	default:
		return errors.New("unknown accrual status")
	}

	if err := order.UpdateStatus(newStatus, accrualResp.Accrual); err != nil {
		return err
	}

	if err := uc.orderRepo.UpdateStatus(ctx, order.ID(), newStatus, accrualResp.Accrual); err != nil {
		return err
	}

	return nil
}

func (uc *ProcessOrdersUseCase) StartWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.InfoContext(ctx, "order processing worker started", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "order processing worker stopped")
			return
		case <-ticker.C:
			if err := uc.ProcessPendingOrders(ctx); err != nil {
				slog.ErrorContext(ctx, "error processing pending orders", "error", err)
			}
		}
	}
}
