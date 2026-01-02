package model

import (
	"errors"
	"time"
)

type Order struct {
	id         int64
	userID     int64
	number     string
	status     OrderStatus
	accrual    *float64
	uploadedAt time.Time
}

func NewOrder(userID int64, number string) (*Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if number == "" {
		return nil, errors.New("order number is required")
	}

	return &Order{
		userID:     userID,
		number:     number,
		status:     OrderStatusNew,
		uploadedAt: time.Now(),
	}, nil
}

func (o *Order) ID() int64 {
	return o.id
}

func (o *Order) UserID() int64 {
	return o.userID
}

func (o *Order) Number() string {
	return o.number
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) Accrual() *float64 {
	return o.accrual
}

func (o *Order) UploadedAt() time.Time {
	return o.uploadedAt
}

func (o *Order) UpdateStatus(newStatus OrderStatus, accrual *float64) error {
	// Бизнес-правило: нельзя перейти из PROCESSED в другой статус
	if o.status == OrderStatusProcessed {
		return errors.New("cannot change status of processed order")
	}

	// Бизнес-правило: только PROCESSED может иметь accrual, и он не может быть отрицательным
	if newStatus == OrderStatusProcessed && accrual != nil && *accrual < 0 {
		return errors.New("accrual cannot be negative")
	}

	// Бизнес-правило: INVALID не может иметь accrual
	if newStatus == OrderStatusInvalid && accrual != nil {
		return errors.New("invalid order cannot have accrual")
	}

	o.status = newStatus
	o.accrual = accrual
	return nil
}

func (o *Order) CanBeUploadedBy(userID int64) bool {
	return o.userID == userID && o.status == OrderStatusNew
}

func (o *Order) IsProcessed() bool {
	return o.status == OrderStatusProcessed
}

func (o *Order) SetID(id int64) {
	o.id = id
}

func RestoreOrder(id, userID int64, number string, status OrderStatus, accrual *float64, uploadedAt time.Time) *Order {
	return &Order{
		id:         id,
		userID:     userID,
		number:     number,
		status:     status,
		accrual:    accrual,
		uploadedAt: uploadedAt,
	}
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)
