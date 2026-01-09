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

func TestOrderRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOrderRepository(pool)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)

		err = repo.Create(ctx, order)

		require.NoError(t, err)
		assert.Greater(t, order.ID(), int64(0), "order ID should be set after creation")

		var dbUserID int64
		var dbNumber string
		err = pool.QueryRow(ctx,
			"SELECT user_id, number FROM orders WHERE id = $1", order.ID(),
		).Scan(&dbUserID, &dbNumber)
		require.NoError(t, err)
		assert.Equal(t, int64(1), dbUserID)
		assert.Equal(t, "79927398713", dbNumber)
	})

	t.Run("duplicate order number", func(t *testing.T) {
		order1, err := model.NewOrder(1, "12345678903")
		require.NoError(t, err)
		err = repo.Create(ctx, order1)
		require.NoError(t, err)

		order2, err := model.NewOrder(2, "12345678903")
		require.NoError(t, err)
		err = repo.Create(ctx, order2)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "order number already exists")
	})
}

func TestOrderRepository_FindByNumber(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOrderRepository(pool)
	ctx := context.Background()

	t.Run("finds existing order", func(t *testing.T) {
		now := time.Now()
		accrual := 100.5
		_, err := pool.Exec(ctx,
			"INSERT INTO orders (user_id, number, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)",
			1, "79927398713", "PROCESSED", accrual, now,
		)
		require.NoError(t, err)

		order, err := repo.FindByNumber(ctx, "79927398713")

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "79927398713", order.Number())
		assert.Equal(t, int64(1), order.UserID())
		assert.Equal(t, model.OrderStatusProcessed, order.Status())
		assert.NotNil(t, order.Accrual())
		assert.Equal(t, accrual, *order.Accrual())
	})

	t.Run("returns nil when order not found", func(t *testing.T) {
		order, err := repo.FindByNumber(ctx, "nonexistent")

		require.NoError(t, err)
		assert.Nil(t, order)
	})
}

func TestOrderRepository_FindByUserID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOrderRepository(pool)
	ctx := context.Background()

	t.Run("returns orders sorted by uploaded_at DESC", func(t *testing.T) {
		now := time.Now()
		_, err := pool.Exec(ctx,
			"INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)",
			1, "order1", "NEW", now.Add(-2*time.Hour),
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)",
			1, "order2", "PROCESSED", now,
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)",
			1, "order3", "NEW", now.Add(-1*time.Hour),
		)
		require.NoError(t, err)

		orders, err := repo.FindByUserID(ctx, 1)

		require.NoError(t, err)
		require.Len(t, orders, 3)
		assert.Equal(t, "order2", orders[0].Number())
		assert.Equal(t, "order3", orders[1].Number())
		assert.Equal(t, "order1", orders[2].Number())
	})

	t.Run("returns empty slice when user has no orders", func(t *testing.T) {
		orders, err := repo.FindByUserID(ctx, 999)

		require.NoError(t, err)
		assert.Len(t, orders, 0)
	})
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOrderRepository(pool)
	ctx := context.Background()

	t.Run("updates status and accrual", func(t *testing.T) {
		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)
		err = repo.Create(ctx, order)
		require.NoError(t, err)

		accrual := 150.75
		err = repo.UpdateStatus(ctx, order.ID(), model.OrderStatusProcessed, &accrual)

		require.NoError(t, err)

		var status string
		var dbAccrual *float64
		err = pool.QueryRow(ctx,
			"SELECT status, accrual FROM orders WHERE id = $1", order.ID(),
		).Scan(&status, &dbAccrual)
		require.NoError(t, err)
		assert.Equal(t, "PROCESSED", status)
		assert.NotNil(t, dbAccrual)
		assert.Equal(t, accrual, *dbAccrual)
	})

	t.Run("updates status with nil accrual", func(t *testing.T) {
		order, err := model.NewOrder(1, "12345678903")
		require.NoError(t, err)
		err = repo.Create(ctx, order)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, order.ID(), model.OrderStatusInvalid, nil)

		require.NoError(t, err)

		var status string
		var dbAccrual *float64
		err = pool.QueryRow(ctx,
			"SELECT status, accrual FROM orders WHERE id = $1", order.ID(),
		).Scan(&status, &dbAccrual)
		require.NoError(t, err)
		assert.Equal(t, "INVALID", status)
		assert.Nil(t, dbAccrual)
	})
}
