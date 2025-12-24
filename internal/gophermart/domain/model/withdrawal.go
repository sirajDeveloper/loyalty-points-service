package model

import (
	"errors"
	"time"
)

type Withdrawal struct {
	id          int64
	userID      int64
	orderNumber string
	sum         float64
	processedAt time.Time
}

func NewWithdrawal(userID int64, orderNumber string, sum float64) (*Withdrawal, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if orderNumber == "" {
		return nil, errors.New("order number is required")
	}
	if sum <= 0 {
		return nil, errors.New("withdrawal sum must be positive")
	}

	return &Withdrawal{
		userID:      userID,
		orderNumber: orderNumber,
		sum:         sum,
		processedAt: time.Now(),
	}, nil
}

func (w *Withdrawal) ID() int64 {
	return w.id
}

func (w *Withdrawal) UserID() int64 {
	return w.userID
}

func (w *Withdrawal) OrderNumber() string {
	return w.orderNumber
}

func (w *Withdrawal) Sum() float64 {
	return w.sum
}

func (w *Withdrawal) ProcessedAt() time.Time {
	return w.processedAt
}

func (w *Withdrawal) SetID(id int64) {
	w.id = id
}

func RestoreWithdrawal(id, userID int64, orderNumber string, sum float64, processedAt time.Time) *Withdrawal {
	return &Withdrawal{
		id:          id,
		userID:      userID,
		orderNumber: orderNumber,
		sum:         sum,
		processedAt: processedAt,
	}
}
