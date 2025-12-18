package model

import (
	"errors"
)

type Balance struct {
	userID    int64
	current   float64
	withdrawn float64
}

func NewBalance(userID int64) *Balance {
	return &Balance{
		userID:    userID,
		current:   0,
		withdrawn: 0,
	}
}

func (b *Balance) UserID() int64 {
	return b.userID
}

func (b *Balance) Current() float64 {
	return b.current
}

func (b *Balance) Withdrawn() float64 {
	return b.withdrawn
}

func (b *Balance) Accrue(amount float64) error {
	if amount <= 0 {
		return errors.New("accrual amount must be positive")
	}
	b.current += amount
	return nil
}

func (b *Balance) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	if b.current < amount {
		return errors.New("insufficient funds")
	}

	b.current -= amount
	b.withdrawn += amount
	return nil
}

func (b *Balance) CanWithdraw(amount float64) bool {
	return amount > 0 && b.current >= amount
}

func RestoreBalance(userID int64, current, withdrawn float64) *Balance {
	return &Balance{
		userID:    userID,
		current:   current,
		withdrawn: withdrawn,
	}
}
