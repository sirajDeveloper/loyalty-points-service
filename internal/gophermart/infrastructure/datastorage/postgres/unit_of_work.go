package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type unitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) repository.UnitOfWork {
	return &unitOfWork{pool: pool}
}

func (uow *unitOfWork) Begin(ctx context.Context) (repository.Transaction, error) {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction{tx: tx}, nil
}

type transaction struct {
	tx pgx.Tx
}

func (t *transaction) OrderRepository() repository.OrderRepository {
	return &orderRepositoryTx{tx: t.tx}
}

func (t *transaction) OutboxRepository() repository.OutboxRepository {
	return &outboxRepositoryTx{tx: t.tx}
}

func (t *transaction) WithdrawalRepository() repository.WithdrawalRepository {
	return &withdrawalRepositoryTx{tx: t.tx}
}

func (t *transaction) BalanceRepository() repository.BalanceRepository {
	return &balanceRepositoryTx{tx: t.tx}
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
