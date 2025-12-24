package repository

import (
	"context"
)

type UnitOfWork interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	OrderRepository() OrderRepository
	OutboxRepository() OutboxRepository
	WithdrawalRepository() WithdrawalRepository
	BalanceRepository() BalanceRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
