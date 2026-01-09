package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 24 * time.Hour

	service := NewJWTService(secretKey, expiry)

	assert.NotNil(t, service)
}

func TestJWTService_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 24 * time.Hour
	service := NewJWTService(secretKey, expiry)

	tests := []struct {
		name   string
		userID int64
		login  string
		check  func(*testing.T, string, error)
	}{
		{
			name:   "successful_token_generation",
			userID: 123,
			login:  "testuser",
			check: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:   "token_with_zero_user_id",
			userID: 0,
			login:  "testuser",
			check: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:   "token_with_empty_login",
			userID: 456,
			login:  "",
			check: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:   "token_with_special_characters_in_login",
			userID: 789,
			login:  "user@example.com",
			check: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:   "token_with_large_user_id",
			userID: 9223372036854775807,
			login:  "testuser",
			check: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.login)
			tt.check(t, token, err)

			if err == nil {
				claims, err := service.ValidateToken(token)
				require.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.login, claims.Login)
			}
		})
	}
}

func TestJWTService_GenerateToken_Claims(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 1 * time.Hour
	service := NewJWTService(secretKey, expiry)

	userID := int64(123)
	login := "testuser"

	token, err := service.GenerateToken(userID, login)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, login, claims.Login)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	now := time.Now()
	expiresAt := claims.ExpiresAt.Time
	issuedAt := claims.IssuedAt.Time

	assert.True(t, expiresAt.After(now), "token should expire in the future")
	assert.True(t, issuedAt.Before(now.Add(1*time.Second)) || issuedAt.Equal(now.Truncate(time.Second)), "token should be issued now or in the past")
	assert.True(t, expiresAt.Sub(issuedAt) >= expiry-time.Second, "token expiry should be approximately equal to configured expiry")
	assert.True(t, expiresAt.Sub(issuedAt) <= expiry+time.Second, "token expiry should be approximately equal to configured expiry")
}

func TestJWTService_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 24 * time.Hour
	service := NewJWTService(secretKey, expiry)

	t.Run("valid_token", func(t *testing.T) {
		userID := int64(123)
		login := "testuser"

		token, err := service.GenerateToken(userID, login)
		require.NoError(t, err)

		claims, err := service.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, login, claims.Login)
	})

	t.Run("invalid_token_format", func(t *testing.T) {
		invalidToken := "invalid.token.format"

		claims, err := service.ValidateToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("empty_token", func(t *testing.T) {
		claims, err := service.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("token_with_wrong_secret", func(t *testing.T) {
		wrongService := NewJWTService("wrong-secret-key", expiry)
		userID := int64(123)
		login := "testuser"

		token, err := wrongService.GenerateToken(userID, login)
		require.NoError(t, err)

		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired_token", func(t *testing.T) {
		shortExpiry := -1 * time.Hour
		expiredService := NewJWTService(secretKey, shortExpiry)

		userID := int64(123)
		login := "testuser"

		token, err := expiredService.GenerateToken(userID, login)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("malformed_token", func(t *testing.T) {
		malformedTokens := []string{
			"not.a.token",
			"header.payload",
			"header.payload.signature.extra",
			"abc",
			".",
			"..",
		}

		for _, malformedToken := range malformedTokens {
			t.Run(malformedToken, func(t *testing.T) {
				claims, err := service.ValidateToken(malformedToken)
				assert.Error(t, err)
				assert.Nil(t, claims)
			})
		}
	})
}

func TestJWTService_GenerateAndValidate_RoundTrip(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 24 * time.Hour
	service := NewJWTService(secretKey, expiry)

	tests := []struct {
		name   string
		userID int64
		login  string
	}{
		{
			name:   "normal_user",
			userID: 123,
			login:  "testuser",
		},
		{
			name:   "user_with_email_login",
			userID: 456,
			login:  "user@example.com",
		},
		{
			name:   "user_with_special_characters",
			userID: 789,
			login:  "user_name-123",
		},
		{
			name:   "zero_user_id",
			userID: 0,
			login:  "testuser",
		},
		{
			name:   "large_user_id",
			userID: 9223372036854775807,
			login:  "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.login)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := service.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.login, claims.Login)
		})
	}
}

func TestJWTService_DifferentExpiryTimes(t *testing.T) {
	secretKey := "test-secret-key"

	expiryTimes := []time.Duration{
		1 * time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
	}

	for _, expiry := range expiryTimes {
		t.Run(expiry.String(), func(t *testing.T) {
			service := NewJWTService(secretKey, expiry)

			userID := int64(123)
			login := "testuser"

			token, err := service.GenerateToken(userID, login)
			require.NoError(t, err)

			claims, err := service.ValidateToken(token)
			require.NoError(t, err)

			now := time.Now()
			expiresAt := claims.ExpiresAt.Time
			issuedAt := claims.IssuedAt.Time

			actualExpiry := expiresAt.Sub(issuedAt)
			assert.True(t, actualExpiry >= expiry-time.Second, "actual expiry should be approximately equal to configured expiry")
			assert.True(t, actualExpiry <= expiry+time.Second, "actual expiry should be approximately equal to configured expiry")
			assert.True(t, expiresAt.After(now), "token should expire in the future")
		})
	}
}

func TestJWTService_DifferentSecretKeys(t *testing.T) {
	expiry := 24 * time.Hour

	secretKeys := []string{
		"short",
		"medium-length-secret-key",
		"very-long-secret-key-that-exceeds-normal-length-requirements",
		"key-with-special-chars!@#$%^&*()",
		"key with spaces",
	}

	for _, secretKey := range secretKeys {
		t.Run(secretKey, func(t *testing.T) {
			service := NewJWTService(secretKey, expiry)

			userID := int64(123)
			login := "testuser"

			token, err := service.GenerateToken(userID, login)
			require.NoError(t, err)

			claims, err := service.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, login, claims.Login)
		})
	}
}
