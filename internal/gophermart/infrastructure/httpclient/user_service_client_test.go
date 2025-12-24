//go:build integration

package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserServiceClient_ValidateToken(t *testing.T) {
	t.Run("successful_validation", func(t *testing.T) {
		expectedUserID := int64(123)
		expectedLogin := "testuser"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/auth/validate", r.URL.Path)

			authHeader := r.Header.Get("Authorization")
			require.NotEmpty(t, authHeader)
			assert.True(t, strings.HasPrefix(authHeader, "Bearer "))

			token := strings.TrimPrefix(authHeader, "Bearer ")
			assert.Equal(t, "valid-token", token)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := ValidateResponse{
				UserID: expectedUserID,
				Login:  expectedLogin,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "valid-token")

		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, expectedUserID, claims.UserID)
		assert.Equal(t, expectedLogin, claims.Login)
	})

	t.Run("invalid_token_401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "invalid-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("unexpected_status_code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "unexpected status code: 500")
	})

	t.Run("invalid_json_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json{"))
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("empty_response_body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("server_unreachable", func(t *testing.T) {
		client := NewUserServiceClient("http://localhost:99999")
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("missing_authorization_header_handled_by_client", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("response_with_zero_values", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := ValidateResponse{
				UserID: 0,
				Login:  "",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, int64(0), claims.UserID)
		assert.Equal(t, "", claims.Login)
	})

	t.Run("response_with_special_characters_in_login", func(t *testing.T) {
		expectedLogin := "user@example.com"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := ValidateResponse{
				UserID: 456,
				Login:  expectedLogin,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		require.NoError(t, err)
		assert.Equal(t, expectedLogin, claims.Login)
	})

	t.Run("large_user_id", func(t *testing.T) {
		expectedUserID := int64(9223372036854775807)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := ValidateResponse{
				UserID: expectedUserID,
				Login:  "testuser",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewUserServiceClient(server.URL)
		ctx := context.Background()

		claims, err := client.ValidateToken(ctx, "some-token")

		require.NoError(t, err)
		assert.Equal(t, expectedUserID, claims.UserID)
	})
}
