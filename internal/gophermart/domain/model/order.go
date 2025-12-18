package model

import "time"

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     OrderStatus
	Accrual    *float64
	UploadedAt time.Time
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

