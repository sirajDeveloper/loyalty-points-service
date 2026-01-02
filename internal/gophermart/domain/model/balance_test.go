package model

import (
	"testing"
)

func TestNewBalance(t *testing.T) {
	userID := int64(123)
	balance := NewBalance(userID)

	if balance == nil {
		t.Fatal("NewBalance() returned nil")
	}

	if balance.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", balance.UserID(), userID)
	}

	if balance.Current() != 0 {
		t.Errorf("Current() = %v, want 0", balance.Current())
	}

	if balance.Withdrawn() != 0 {
		t.Errorf("Withdrawn() = %v, want 0", balance.Withdrawn())
	}
}

func TestRestoreBalance(t *testing.T) {
	userID := int64(123)
	current := 100.5
	withdrawn := 50.0

	balance := RestoreBalance(userID, current, withdrawn)

	if balance.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", balance.UserID(), userID)
	}

	if balance.Current() != current {
		t.Errorf("Current() = %v, want %v", balance.Current(), current)
	}

	if balance.Withdrawn() != withdrawn {
		t.Errorf("Withdrawn() = %v, want %v", balance.Withdrawn(), withdrawn)
	}
}

func TestBalance_Accrue(t *testing.T) {
	balance := NewBalance(1)

	tests := []struct {
		name        string
		amount      float64
		wantErr     bool
		wantCurrent float64
	}{
		{"positive amount", 100.0, false, 100.0},
		{"another positive amount", 50.5, false, 150.5},
		{"zero amount", 0.0, true, 150.5},
		{"negative amount", -10.0, true, 150.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := balance.Accrue(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accrue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && balance.Current() != tt.wantCurrent {
				t.Errorf("Current() = %v, want %v", balance.Current(), tt.wantCurrent)
			}
		})
	}
}

func TestBalance_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		initial       float64
		amount        float64
		wantErr       bool
		wantCurrent   float64
		wantWithdrawn float64
	}{
		{"successful withdrawal", 100.0, 50.0, false, 50.0, 50.0},
		{"withdraw all", 100.0, 100.0, false, 0.0, 100.0},
		{"insufficient funds", 50.0, 100.0, true, 50.0, 0.0},
		{"zero amount", 100.0, 0.0, true, 100.0, 0.0},
		{"negative amount", 100.0, -10.0, true, 100.0, 0.0},
		{"multiple withdrawals", 100.0, 30.0, false, 70.0, 30.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance := RestoreBalance(1, tt.initial, 0.0)
			err := balance.Withdraw(tt.amount)

			if (err != nil) != tt.wantErr {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if balance.Current() != tt.wantCurrent {
					t.Errorf("Current() = %v, want %v", balance.Current(), tt.wantCurrent)
				}
				if balance.Withdrawn() != tt.wantWithdrawn {
					t.Errorf("Withdrawn() = %v, want %v", balance.Withdrawn(), tt.wantWithdrawn)
				}
			} else {
				if balance.Current() != tt.initial {
					t.Errorf("Current() should not change on error, got %v, want %v", balance.Current(), tt.initial)
				}
			}
		})
	}
}

func TestBalance_CanWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		current float64
		amount  float64
		want    bool
	}{
		{"can withdraw", 100.0, 50.0, true},
		{"can withdraw all", 100.0, 100.0, true},
		{"cannot withdraw more", 50.0, 100.0, false},
		{"cannot withdraw zero", 100.0, 0.0, false},
		{"cannot withdraw negative", 100.0, -10.0, false},
		{"exact amount", 100.0, 100.0, true},
		{"small amount", 0.01, 0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance := RestoreBalance(1, tt.current, 0.0)
			if got := balance.CanWithdraw(tt.amount); got != tt.want {
				t.Errorf("CanWithdraw(%v) = %v, want %v", tt.amount, got, tt.want)
			}
		})
	}
}

func TestBalance_Current(t *testing.T) {
	balance := RestoreBalance(1, 150.75, 0.0)
	if balance.Current() != 150.75 {
		t.Errorf("Current() = %v, want 150.75", balance.Current())
	}
}

func TestBalance_Withdrawn(t *testing.T) {
	balance := RestoreBalance(1, 100.0, 75.5)
	if balance.Withdrawn() != 75.5 {
		t.Errorf("Withdrawn() = %v, want 75.5", balance.Withdrawn())
	}
}

func TestBalance_UserID(t *testing.T) {
	userID := int64(999)
	balance := NewBalance(userID)
	if balance.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", balance.UserID(), userID)
	}
}
