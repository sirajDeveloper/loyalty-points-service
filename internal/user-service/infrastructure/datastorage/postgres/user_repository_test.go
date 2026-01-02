//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/datastorage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	setupTestDB(t)
	repo := postgres.NewUserRepository(testPool)
	ctx := context.Background()

	t.Run("creates_user_successfully", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "testuser",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("returns_error_on_duplicate_login", func(t *testing.T) {
		setupTestDB(t)
		user1 := &model.User{
			Login:        "duplicate_login",
			PasswordHash: "hash1",
			FirstName:    "First",
			LastName:     "User",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &model.User{
			Login:        "duplicate_login",
			PasswordHash: "hash2",
			FirstName:    "Second",
			LastName:     "User",
			CreatedAt:    time.Now(),
		}
		err = repo.Create(ctx, user2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrLoginAlreadyExists))
	})

	t.Run("creates_user_with_empty_first_name", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "user_no_firstname",
			PasswordHash: "hash",
			FirstName:    "",
			LastName:     "Doe",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("creates_user_with_empty_last_name", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "user_no_lastname",
			PasswordHash: "hash",
			FirstName:    "John",
			LastName:     "",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("creates_user_with_email_as_login", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "user@example.com",
			PasswordHash: "hash",
			FirstName:    "John",
			LastName:     "Doe",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})
}

func TestUserRepository_FindByLogin(t *testing.T) {
	setupTestDB(t)
	repo := postgres.NewUserRepository(testPool)
	ctx := context.Background()

	t.Run("returns_user_when_exists", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "finduser",
			PasswordHash: "hashed_password",
			FirstName:    "Find",
			LastName:     "User",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByLogin(ctx, "finduser")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "finduser", found.Login)
		assert.Equal(t, "hashed_password", found.PasswordHash)
		assert.Equal(t, "Find", found.FirstName)
		assert.Equal(t, "User", found.LastName)
	})

	t.Run("returns_error_when_user_not_found", func(t *testing.T) {
		setupTestDB(t)
		found, err := repo.FindByLogin(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.True(t, errors.Is(err, errors.ErrUserNotFound))
	})

	t.Run("finds_user_with_email_login", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "email@example.com",
			PasswordHash: "hash",
			FirstName:    "Email",
			LastName:     "User",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByLogin(ctx, "email@example.com")
		require.NoError(t, err)
		assert.Equal(t, "email@example.com", found.Login)
	})

	t.Run("case_sensitive_login", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "CaseSensitive",
			PasswordHash: "hash",
			FirstName:    "Case",
			LastName:     "Sensitive",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByLogin(ctx, "casesensitive")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.True(t, errors.Is(err, errors.ErrUserNotFound))
	})
}

func TestUserRepository_ExistsByLogin(t *testing.T) {
	setupTestDB(t)
	repo := postgres.NewUserRepository(testPool)
	ctx := context.Background()

	t.Run("returns_true_when_user_exists", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "existsuser",
			PasswordHash: "hash",
			FirstName:    "Exists",
			LastName:     "User",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := repo.ExistsByLogin(ctx, "existsuser")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns_false_when_user_not_exists", func(t *testing.T) {
		setupTestDB(t)
		exists, err := repo.ExistsByLogin(ctx, "nonexistent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns_false_for_empty_login", func(t *testing.T) {
		setupTestDB(t)
		exists, err := repo.ExistsByLogin(ctx, "")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("case_sensitive_check", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "CaseSensitive",
			PasswordHash: "hash",
			FirstName:    "Case",
			LastName:     "Sensitive",
			CreatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := repo.ExistsByLogin(ctx, "casesensitive")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestUserRepository_RoundTrip(t *testing.T) {
	setupTestDB(t)
	repo := postgres.NewUserRepository(testPool)
	ctx := context.Background()

	t.Run("create_and_find_user", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "roundtrip",
			PasswordHash: "hashed_password_123",
			FirstName:    "Round",
			LastName:     "Trip",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByLogin(ctx, "roundtrip")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Login, found.Login)
		assert.Equal(t, user.PasswordHash, found.PasswordHash)
		assert.Equal(t, user.FirstName, found.FirstName)
		assert.Equal(t, user.LastName, found.LastName)
	})

	t.Run("create_and_check_exists", func(t *testing.T) {
		setupTestDB(t)
		user := &model.User{
			Login:        "checkexists",
			PasswordHash: "hash",
			FirstName:    "Check",
			LastName:     "Exists",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := repo.ExistsByLogin(ctx, "checkexists")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}
