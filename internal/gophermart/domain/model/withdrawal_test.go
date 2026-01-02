package model

import (
	"testing"
	"time"
)

func TestNewWithdrawal(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		orderNumber string
		sum         float64
		wantErr     bool
	}{
		{"valid withdrawal", 1, "12345678903", 100.0, false},
		{"another valid", 999, "79927398713", 50.5, false},
		{"invalid user ID zero", 0, "12345678903", 100.0, true},
		{"invalid user ID negative", -1, "12345678903", 100.0, true},
		{"empty order number", 1, "", 100.0, true},
		{"zero sum", 1, "12345678903", 0.0, true},
		{"negative sum", 1, "12345678903", -10.0, true},
		{"small positive sum", 1, "12345678903", 0.01, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithdrawal(tt.userID, tt.orderNumber, tt.sum)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithdrawal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Fatal("NewWithdrawal() returned nil")
				}
				if got.UserID() != tt.userID {
					t.Errorf("UserID() = %v, want %v", got.UserID(), tt.userID)
				}
				if got.OrderNumber() != tt.orderNumber {
					t.Errorf("OrderNumber() = %v, want %v", got.OrderNumber(), tt.orderNumber)
				}
				if got.Sum() != tt.sum {
					t.Errorf("Sum() = %v, want %v", got.Sum(), tt.sum)
				}
			}
		})
	}
}

func TestRestoreWithdrawal(t *testing.T) {
	id := int64(1)
	userID := int64(123)
	orderNumber := "12345678903"
	sum := 150.75
	processedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	withdrawal := RestoreWithdrawal(id, userID, orderNumber, sum, processedAt)

	if withdrawal.ID() != id {
		t.Errorf("ID() = %v, want %v", withdrawal.ID(), id)
	}
	if withdrawal.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", withdrawal.UserID(), userID)
	}
	if withdrawal.OrderNumber() != orderNumber {
		t.Errorf("OrderNumber() = %v, want %v", withdrawal.OrderNumber(), orderNumber)
	}
	if withdrawal.Sum() != sum {
		t.Errorf("Sum() = %v, want %v", withdrawal.Sum(), sum)
	}
	if !withdrawal.ProcessedAt().Equal(processedAt) {
		t.Errorf("ProcessedAt() = %v, want %v", withdrawal.ProcessedAt(), processedAt)
	}
}

func TestWithdrawal_SetID(t *testing.T) {
	withdrawal := RestoreWithdrawal(1, 1, "12345678903", 100.0, time.Now())
	newID := int64(999)

	withdrawal.SetID(newID)

	if withdrawal.ID() != newID {
		t.Errorf("ID() = %v, want %v", withdrawal.ID(), newID)
	}
}

func TestWithdrawal_ID(t *testing.T) {
	id := int64(123)
	withdrawal := RestoreWithdrawal(id, 1, "12345678903", 100.0, time.Now())
	if withdrawal.ID() != id {
		t.Errorf("ID() = %v, want %v", withdrawal.ID(), id)
	}
}

func TestWithdrawal_UserID(t *testing.T) {
	userID := int64(456)
	withdrawal := RestoreWithdrawal(1, userID, "12345678903", 100.0, time.Now())
	if withdrawal.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", withdrawal.UserID(), userID)
	}
}

func TestWithdrawal_OrderNumber(t *testing.T) {
	orderNumber := "79927398713"
	withdrawal := RestoreWithdrawal(1, 1, orderNumber, 100.0, time.Now())
	if withdrawal.OrderNumber() != orderNumber {
		t.Errorf("OrderNumber() = %v, want %v", withdrawal.OrderNumber(), orderNumber)
	}
}

func TestWithdrawal_Sum(t *testing.T) {
	sum := 250.5
	withdrawal := RestoreWithdrawal(1, 1, "12345678903", sum, time.Now())
	if withdrawal.Sum() != sum {
		t.Errorf("Sum() = %v, want %v", withdrawal.Sum(), sum)
	}
}

func TestWithdrawal_ProcessedAt(t *testing.T) {
	processedAt := time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC)
	withdrawal := RestoreWithdrawal(1, 1, "12345678903", 100.0, processedAt)
	if !withdrawal.ProcessedAt().Equal(processedAt) {
		t.Errorf("ProcessedAt() = %v, want %v", withdrawal.ProcessedAt(), processedAt)
	}
}
