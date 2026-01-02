//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/datastorage/postgres"
)

func TestBalanceRepository_GetByUserID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBalanceRepository(pool)
	ctx := context.Background()

	t.Run("returns new balance when user not found", func(t *testing.T) {
		balance, err := repo.GetByUserID(ctx, 999)

		require.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, int64(999), balance.UserID())
		assert.Equal(t, 0.0, balance.Current())
		assert.Equal(t, 0.0, balance.Withdrawn())
	})

	t.Run("returns existing balance", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)",
			1, 100.5, 50.0,
		)
		require.NoError(t, err)

		balance, err := repo.GetByUserID(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, int64(1), balance.UserID())
		assert.Equal(t, 100.5, balance.Current())
		assert.Equal(t, 50.0, balance.Withdrawn())
	})
}

func TestBalanceRepository_Withdraw(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBalanceRepository(pool)
	ctx := context.Background()

	t.Run("successful withdrawal from existing balance", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)",
			1, 100.0, 0.0,
		)
		require.NoError(t, err)

		err = repo.Withdraw(ctx, 1, 50.0)

		require.NoError(t, err)

		var current, withdrawn float64
		err = pool.QueryRow(ctx,
			"SELECT current, withdrawn FROM balances WHERE user_id = $1", 1,
		).Scan(&current, &withdrawn)
		require.NoError(t, err)

		assert.Equal(t, 50.0, current)
		assert.Equal(t, 50.0, withdrawn)
	})

	t.Run("creates balance if not exists", func(t *testing.T) {
		err := repo.Withdraw(ctx, 2, 30.0)

		require.NoError(t, err)

		var current, withdrawn float64
		err = pool.QueryRow(ctx,
			"SELECT current, withdrawn FROM balances WHERE user_id = $1", 2,
		).Scan(&current, &withdrawn)
		require.NoError(t, err)

		assert.Equal(t, -30.0, current)
		assert.Equal(t, 30.0, withdrawn)
	})

	t.Run("multiple withdrawals accumulate", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)",
			3, 200.0, 10.0,
		)
		require.NoError(t, err)

		err = repo.Withdraw(ctx, 3, 75.0)
		require.NoError(t, err)

		var current, withdrawn float64
		err = pool.QueryRow(ctx,
			"SELECT current, withdrawn FROM balances WHERE user_id = $1", 3,
		).Scan(&current, &withdrawn)
		require.NoError(t, err)

		assert.Equal(t, 125.0, current)
		assert.Equal(t, 85.0, withdrawn)
	})
}

func TestBalanceRepository_Accrue(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewBalanceRepository(pool)
	ctx := context.Background()

	t.Run("successful accrue to existing balance", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)",
			1, 100.0, 0.0,
		)
		require.NoError(t, err)

		err = repo.Accrue(ctx, 1, 50.0)

		require.NoError(t, err)

		var current float64
		err = pool.QueryRow(ctx,
			"SELECT current FROM balances WHERE user_id = $1", 1,
		).Scan(&current)
		require.NoError(t, err)

		assert.Equal(t, 150.0, current)
	})

	t.Run("creates balance if not exists", func(t *testing.T) {
		err := repo.Accrue(ctx, 2, 75.5)

		require.NoError(t, err)

		var current float64
		err = pool.QueryRow(ctx,
			"SELECT current FROM balances WHERE user_id = $1", 2,
		).Scan(&current)
		require.NoError(t, err)

		assert.Equal(t, 75.5, current)
	})

	t.Run("multiple accruals accumulate", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			"INSERT INTO balances (user_id, current, withdrawn) VALUES ($1, $2, $3)",
			3, 50.0, 0.0,
		)
		require.NoError(t, err)

		err = repo.Accrue(ctx, 3, 25.0)
		require.NoError(t, err)
		err = repo.Accrue(ctx, 3, 25.0)
		require.NoError(t, err)

		var current float64
		err = pool.QueryRow(ctx,
			"SELECT current FROM balances WHERE user_id = $1", 3,
		).Scan(&current)
		require.NoError(t, err)

		assert.Equal(t, 100.0, current)
	})
}
