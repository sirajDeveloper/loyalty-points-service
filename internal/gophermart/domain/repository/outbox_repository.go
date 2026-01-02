package repository

import (
	"context"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

type OutboxRepository interface {
	Create(ctx context.Context, outbox *model.Outbox) error
	FindPending(ctx context.Context, limit int) ([]*model.Outbox, error)
	UpdateStatus(ctx context.Context, outboxID int64, status model.OutboxStatus) error
	IncrementRetries(ctx context.Context, outboxID int64) error
}


