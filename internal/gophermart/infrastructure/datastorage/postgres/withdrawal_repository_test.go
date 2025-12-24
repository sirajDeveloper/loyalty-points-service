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

func TestWithdrawalRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewWithdrawalRepository(pool)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		withdrawal, err := model.NewWithdrawal(1, "79927398713", 100.5)
		require.NoError(t, err)

		err = repo.Create(ctx, withdrawal)

		require.NoError(t, err)
		assert.Greater(t, withdrawal.ID(), int64(0))

		var dbUserID int64
		var dbOrderNumber string
		var dbSum float64
		err = pool.QueryRow(ctx,
			"SELECT user_id, order_number, sum FROM withdrawals WHERE id = $1", withdrawal.ID(),
		).Scan(&dbUserID, &dbOrderNumber, &dbSum)
		require.NoError(t, err)
		assert.Equal(t, int64(1), dbUserID)
		assert.Equal(t, "79927398713", dbOrderNumber)
		assert.Equal(t, 100.5, dbSum)
	})
}

func TestWithdrawalRepository_FindByUserID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewWithdrawalRepository(pool)
	ctx := context.Background()

	t.Run("returns withdrawals sorted by processed_at DESC", func(t *testing.T) {
		now := time.Now()
		_, err := pool.Exec(ctx,
			"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
			1, "order1", 50.0, now.Add(-2*time.Hour),
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
			1, "order2", 75.0, now,
		)
		require.NoError(t, err)
		_, err = pool.Exec(ctx,
			"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
			1, "order3", 25.0, now.Add(-1*time.Hour),
		)
		require.NoError(t, err)

		withdrawals, err := repo.FindByUserID(ctx, 1)

		require.NoError(t, err)
		require.Len(t, withdrawals, 3)
		assert.Equal(t, "order2", withdrawals[0].OrderNumber())
		assert.Equal(t, 75.0, withdrawals[0].Sum())
		assert.Equal(t, "order3", withdrawals[1].OrderNumber())
		assert.Equal(t, 25.0, withdrawals[1].Sum())
		assert.Equal(t, "order1", withdrawals[2].OrderNumber())
		assert.Equal(t, 50.0, withdrawals[2].Sum())
	})

	t.Run("returns empty slice when user has no withdrawals", func(t *testing.T) {
		withdrawals, err := repo.FindByUserID(ctx, 999)

		require.NoError(t, err)
		assert.Len(t, withdrawals, 0)
	})
}
