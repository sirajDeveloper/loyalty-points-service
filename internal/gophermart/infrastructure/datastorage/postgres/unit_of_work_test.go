//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/datastorage/postgres"
)

func TestUnitOfWork_BeginCommit(t *testing.T) {
	pool := setupTestDB(t)
	uow := postgres.NewUnitOfWork(pool)
	ctx := context.Background()

	t.Run("successful transaction commits all changes", func(t *testing.T) {
		_, err := pool.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", "user1", "hash")
		require.NoError(t, err)

		tx, err := uow.Begin(ctx)
		require.NoError(t, err)

		balanceRepo := tx.BalanceRepository()
		err = balanceRepo.Accrue(ctx, 1, 100.0)
		require.NoError(t, err)

		orderRepo := tx.OrderRepository()
		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)
		err = orderRepo.Create(ctx, order)
		require.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)

		var balance float64
		err = pool.QueryRow(ctx,
			"SELECT current FROM balances WHERE user_id = $1", 1,
		).Scan(&balance)
		require.NoError(t, err)
		assert.Equal(t, 100.0, balance)

		var orderExists bool
		err = pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)", "79927398713",
		).Scan(&orderExists)
		require.NoError(t, err)
		assert.True(t, orderExists)
	})
}

func TestUnitOfWork_BeginRollback(t *testing.T) {
	pool := setupTestDB(t)
	uow := postgres.NewUnitOfWork(pool)
	ctx := context.Background()

	t.Run("rollback discards all changes", func(t *testing.T) {
		_, err := pool.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", "user1", "hash")
		require.NoError(t, err)

		tx, err := uow.Begin(ctx)
		require.NoError(t, err)

		balanceRepo := tx.BalanceRepository()
		err = balanceRepo.Accrue(ctx, 1, 100.0)
		require.NoError(t, err)

		orderRepo := tx.OrderRepository()
		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)
		err = orderRepo.Create(ctx, order)
		require.NoError(t, err)

		err = tx.Rollback(ctx)
		require.NoError(t, err)

		var count int
		err = pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM balances WHERE user_id = $1", 1,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		err = pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM orders WHERE number = $1", "79927398713",
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestUnitOfWork_TransactionIsolation(t *testing.T) {
	pool := setupTestDB(t)
	uow := postgres.NewUnitOfWork(pool)
	ctx := context.Background()

	t.Run("changes in transaction are not visible until commit", func(t *testing.T) {
		_, err := pool.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", "user1", "hash")
		require.NoError(t, err)

		tx1, err := uow.Begin(ctx)
		require.NoError(t, err)

		balanceRepo1 := tx1.BalanceRepository()
		err = balanceRepo1.Accrue(ctx, 1, 100.0)
		require.NoError(t, err)

		balance, err := balanceRepo1.GetByUserID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 100.0, balance.Current())

		balanceRepo2 := postgres.NewBalanceRepository(pool)
		balance2, err := balanceRepo2.GetByUserID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance2.Current())

		err = tx1.Commit(ctx)
		require.NoError(t, err)

		balance3, err := balanceRepo2.GetByUserID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 100.0, balance3.Current())
	})
}

func TestUnitOfWork_TransactionRepositories(t *testing.T) {
	pool := setupTestDB(t)
	uow := postgres.NewUnitOfWork(pool)
	ctx := context.Background()

	t.Run("all repository types are available in transaction", func(t *testing.T) {
		_, err := pool.Exec(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", "user1", "hash")
		require.NoError(t, err)

		tx, err := uow.Begin(ctx)
		require.NoError(t, err)

		orderRepo := tx.OrderRepository()
		assert.NotNil(t, orderRepo)

		outboxRepo := tx.OutboxRepository()
		assert.NotNil(t, outboxRepo)

		balanceRepo := tx.BalanceRepository()
		assert.NotNil(t, balanceRepo)

		withdrawalRepo := tx.WithdrawalRepository()
		assert.NotNil(t, withdrawalRepo)

		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)
		err = orderRepo.Create(ctx, order)
		require.NoError(t, err)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = outboxRepo.Create(ctx, outbox)
		require.NoError(t, err)

		err = balanceRepo.Accrue(ctx, 1, 50.0)
		require.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)

		var orderCount, outboxCount int
		var balance float64

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE number = $1", "79927398713").Scan(&orderCount)
		require.NoError(t, err)
		assert.Equal(t, 1, orderCount)

		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE order_id = $1", order.ID()).Scan(&outboxCount)
		require.NoError(t, err)
		assert.Equal(t, 1, outboxCount)

		err = pool.QueryRow(ctx, "SELECT current FROM balances WHERE user_id = $1", 1).Scan(&balance)
		require.NoError(t, err)
		assert.Equal(t, 50.0, balance)
	})
}
