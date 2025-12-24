//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/datastorage/postgres"
)

func TestOutboxRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOutboxRepository(pool)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		order, err := model.NewOrder(1, "79927398713")
		require.NoError(t, err)
		orderRepo := postgres.NewOrderRepository(pool)
		err = orderRepo.Create(ctx, order)
		require.NoError(t, err)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = repo.Create(ctx, outbox)

		require.NoError(t, err)
		assert.Greater(t, outbox.ID, int64(0))

		var dbOrderID int64
		var dbStatus string
		var dbRetries int
		err = pool.QueryRow(ctx,
			"SELECT order_id, status, retries FROM outbox WHERE id = $1", outbox.ID,
		).Scan(&dbOrderID, &dbStatus, &dbRetries)
		require.NoError(t, err)
		assert.Equal(t, order.ID(), dbOrderID)
		assert.Equal(t, "PENDING", dbStatus)
		assert.Equal(t, 0, dbRetries)
	})
}

func TestOutboxRepository_FindPending(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOutboxRepository(pool)
	ctx := context.Background()

	t.Run("returns only pending outboxes sorted by created_at ASC", func(t *testing.T) {
		orderRepo := postgres.NewOrderRepository(pool)

		order1, _ := model.NewOrder(1, "order1")
		orderRepo.Create(ctx, order1)
		order2, _ := model.NewOrder(1, "order2")
		orderRepo.Create(ctx, order2)
		order3, _ := model.NewOrder(1, "order3")
		orderRepo.Create(ctx, order3)

		now := time.Now()
		_, err := pool.Exec(ctx,
			"INSERT INTO outbox (order_id, status, retries, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			order1.ID(), "PENDING", 0, now.Add(-2*time.Hour), now,
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO outbox (order_id, status, retries, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			order2.ID(), "PROCESSED", 0, now.Add(-1*time.Hour), now,
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO outbox (order_id, status, retries, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			order3.ID(), "PENDING", 0, now.Add(-1*time.Hour), now,
		)
		require.NoError(t, err)

		outboxes, err := repo.FindPending(ctx, 10)

		require.NoError(t, err)
		require.Len(t, outboxes, 2)
		assert.Equal(t, order1.ID(), outboxes[0].OrderID)
		assert.Equal(t, model.OutboxStatusPending, outboxes[0].Status)
		assert.Equal(t, order3.ID(), outboxes[1].OrderID)
		assert.Equal(t, model.OutboxStatusPending, outboxes[1].Status)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		orderRepo := postgres.NewOrderRepository(pool)
		for i := 0; i < 5; i++ {
			order, _ := model.NewOrder(1, fmt.Sprintf("order%d", i))
			orderRepo.Create(ctx, order)
			outbox := &model.Outbox{
				OrderID:   order.ID(),
				Status:    model.OutboxStatusPending,
				Retries:   0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			repo.Create(ctx, outbox)
		}

		outboxes, err := repo.FindPending(ctx, 3)

		require.NoError(t, err)
		assert.Len(t, outboxes, 3)
	})
}

func TestOutboxRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOutboxRepository(pool)
	ctx := context.Background()

	t.Run("updates status to PROCESSED", func(t *testing.T) {
		order, _ := model.NewOrder(1, "79927398713")
		orderRepo := postgres.NewOrderRepository(pool)
		orderRepo.Create(ctx, order)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, outbox)

		err := repo.UpdateStatus(ctx, outbox.ID, model.OutboxStatusProcessed)

		require.NoError(t, err)

		var status string
		err = pool.QueryRow(ctx,
			"SELECT status FROM outbox WHERE id = $1", outbox.ID,
		).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "PROCESSED", status)
	})

	t.Run("updates status to FAILED", func(t *testing.T) {
		order, _ := model.NewOrder(1, "12345678903")
		orderRepo := postgres.NewOrderRepository(pool)
		orderRepo.Create(ctx, order)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, outbox)

		err := repo.UpdateStatus(ctx, outbox.ID, model.OutboxStatusFailed)

		require.NoError(t, err)

		var status string
		err = pool.QueryRow(ctx,
			"SELECT status FROM outbox WHERE id = $1", outbox.ID,
		).Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "FAILED", status)
	})
}

func TestOutboxRepository_IncrementRetries(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewOutboxRepository(pool)
	ctx := context.Background()

	t.Run("increments retries counter", func(t *testing.T) {
		order, _ := model.NewOrder(1, "79927398713")
		orderRepo := postgres.NewOrderRepository(pool)
		orderRepo.Create(ctx, order)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, outbox)

		err := repo.IncrementRetries(ctx, outbox.ID)

		require.NoError(t, err)

		var retries int
		err = pool.QueryRow(ctx,
			"SELECT retries FROM outbox WHERE id = $1", outbox.ID,
		).Scan(&retries)
		require.NoError(t, err)
		assert.Equal(t, 2, retries)
	})

	t.Run("multiple increments work correctly", func(t *testing.T) {
		order, _ := model.NewOrder(1, "12345678903")
		orderRepo := postgres.NewOrderRepository(pool)
		orderRepo.Create(ctx, order)

		outbox := &model.Outbox{
			OrderID:   order.ID(),
			Status:    model.OutboxStatusPending,
			Retries:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(ctx, outbox)

		err := repo.IncrementRetries(ctx, outbox.ID)
		require.NoError(t, err)
		err = repo.IncrementRetries(ctx, outbox.ID)
		require.NoError(t, err)

		var retries int
		err = pool.QueryRow(ctx,
			"SELECT retries FROM outbox WHERE id = $1", outbox.ID,
		).Scan(&retries)
		require.NoError(t, err)
		assert.Equal(t, 2, retries)
	})
}
