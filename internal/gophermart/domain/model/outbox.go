package model

import "time"

type Outbox struct {
	ID        int64
	OrderID   int64
	Status    OutboxStatus
	Retries   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OutboxStatus string

const (
	OutboxStatusPending  OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)


