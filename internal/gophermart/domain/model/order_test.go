package model

import (
	"testing"
	"time"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		number  string
		wantErr bool
	}{
		{"valid order", 1, "12345678903", false},
		{"another valid order", 999, "79927398713", false},
		{"invalid user ID zero", 0, "12345678903", true},
		{"invalid user ID negative", -1, "12345678903", true},
		{"empty number", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOrder(tt.userID, tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Fatal("NewOrder() returned nil")
				}
				if got.UserID() != tt.userID {
					t.Errorf("UserID() = %v, want %v", got.UserID(), tt.userID)
				}
				if got.Number() != tt.number {
					t.Errorf("Number() = %v, want %v", got.Number(), tt.number)
				}
				if got.Status() != OrderStatusNew {
					t.Errorf("Status() = %v, want %v", got.Status(), OrderStatusNew)
				}
				if got.Accrual() != nil {
					t.Errorf("Accrual() = %v, want nil", got.Accrual())
				}
			}
		})
	}
}

func TestRestoreOrder(t *testing.T) {
	id := int64(1)
	userID := int64(123)
	number := "12345678903"
	status := OrderStatusProcessed
	accrual := 100.5
	uploadedAt := time.Now()

	order := RestoreOrder(id, userID, number, status, &accrual, uploadedAt)

	if order.ID() != id {
		t.Errorf("ID() = %v, want %v", order.ID(), id)
	}
	if order.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", order.UserID(), userID)
	}
	if order.Number() != number {
		t.Errorf("Number() = %v, want %v", order.Number(), number)
	}
	if order.Status() != status {
		t.Errorf("Status() = %v, want %v", order.Status(), status)
	}
	if order.Accrual() == nil || *order.Accrual() != accrual {
		t.Errorf("Accrual() = %v, want %v", order.Accrual(), accrual)
	}
	if !order.UploadedAt().Equal(uploadedAt) {
		t.Errorf("UploadedAt() = %v, want %v", order.UploadedAt(), uploadedAt)
	}
}

func TestOrder_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		newStatus     OrderStatus
		accrual       *float64
		wantErr       bool
	}{
		{"new to processing", OrderStatusNew, OrderStatusProcessing, nil, false},
		{"processing to processed", OrderStatusProcessing, OrderStatusProcessed, floatPtr(100.0), false},
		{"processing to invalid", OrderStatusProcessing, OrderStatusInvalid, nil, false},
		{"cannot change processed", OrderStatusProcessed, OrderStatusNew, nil, true},
		{"processed to processing", OrderStatusProcessed, OrderStatusProcessing, nil, true},
		{"negative accrual", OrderStatusProcessing, OrderStatusProcessed, floatPtr(-10.0), true},
		{"invalid with accrual", OrderStatusProcessing, OrderStatusInvalid, floatPtr(100.0), true},
		{"processed with nil accrual", OrderStatusProcessing, OrderStatusProcessed, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := RestoreOrder(1, 1, "12345678903", tt.initialStatus, nil, time.Now())
			err := order.UpdateStatus(tt.newStatus, tt.accrual)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if order.Status() != tt.newStatus {
					t.Errorf("Status() = %v, want %v", order.Status(), tt.newStatus)
				}
				if tt.accrual != nil {
					if order.Accrual() == nil || *order.Accrual() != *tt.accrual {
						t.Errorf("Accrual() = %v, want %v", order.Accrual(), tt.accrual)
					}
				}
			}
		})
	}
}

func TestOrder_CanBeUploadedBy(t *testing.T) {
	tests := []struct {
		name        string
		orderUserID int64
		status      OrderStatus
		userID      int64
		want        bool
	}{
		{"same user new status", 1, OrderStatusNew, 1, true},
		{"different user new status", 1, OrderStatusNew, 2, false},
		{"same user processing", 1, OrderStatusProcessing, 1, false},
		{"same user processed", 1, OrderStatusProcessed, 1, false},
		{"same user invalid", 1, OrderStatusInvalid, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := RestoreOrder(1, tt.orderUserID, "12345678903", tt.status, nil, time.Now())
			if got := order.CanBeUploadedBy(tt.userID); got != tt.want {
				t.Errorf("CanBeUploadedBy(%v) = %v, want %v", tt.userID, got, tt.want)
			}
		})
	}
}

func TestOrder_IsProcessed(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"processed", OrderStatusProcessed, true},
		{"new", OrderStatusNew, false},
		{"processing", OrderStatusProcessing, false},
		{"invalid", OrderStatusInvalid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := RestoreOrder(1, 1, "12345678903", tt.status, nil, time.Now())
			if got := order.IsProcessed(); got != tt.want {
				t.Errorf("IsProcessed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_SetID(t *testing.T) {
	newID := int64(999)

	order := RestoreOrder(1, 1, "12345678903", OrderStatusNew, nil, time.Now())
	order.SetID(newID)

	if order.ID() != newID {
		t.Errorf("ID() = %v, want %v", order.ID(), newID)
	}
}

func TestOrder_ID(t *testing.T) {
	id := int64(123)
	order := RestoreOrder(id, 1, "12345678903", OrderStatusNew, nil, time.Now())
	if order.ID() != id {
		t.Errorf("ID() = %v, want %v", order.ID(), id)
	}
}

func TestOrder_UserID(t *testing.T) {
	userID := int64(456)
	order := RestoreOrder(1, userID, "12345678903", OrderStatusNew, nil, time.Now())
	if order.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", order.UserID(), userID)
	}
}

func TestOrder_Number(t *testing.T) {
	number := "79927398713"
	order := RestoreOrder(1, 1, number, OrderStatusNew, nil, time.Now())
	if order.Number() != number {
		t.Errorf("Number() = %v, want %v", order.Number(), number)
	}
}

func TestOrder_Status(t *testing.T) {
	status := OrderStatusProcessing
	order := RestoreOrder(1, 1, "12345678903", status, nil, time.Now())
	if order.Status() != status {
		t.Errorf("Status() = %v, want %v", order.Status(), status)
	}
}

func TestOrder_Accrual(t *testing.T) {
	accrual := 150.75
	order := RestoreOrder(1, 1, "12345678903", OrderStatusProcessed, &accrual, time.Now())
	if order.Accrual() == nil || *order.Accrual() != accrual {
		t.Errorf("Accrual() = %v, want %v", order.Accrual(), accrual)
	}

	orderNoAccrual := RestoreOrder(1, 1, "12345678903", OrderStatusNew, nil, time.Now())
	if orderNoAccrual.Accrual() != nil {
		t.Errorf("Accrual() = %v, want nil", orderNoAccrual.Accrual())
	}
}

func TestOrder_UploadedAt(t *testing.T) {
	uploadedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	order := RestoreOrder(1, 1, "12345678903", OrderStatusNew, nil, uploadedAt)
	if !order.UploadedAt().Equal(uploadedAt) {
		t.Errorf("UploadedAt() = %v, want %v", order.UploadedAt(), uploadedAt)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
