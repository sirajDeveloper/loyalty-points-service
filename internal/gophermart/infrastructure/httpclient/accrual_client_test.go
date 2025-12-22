package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccrualClient_GetOrderInfo(t *testing.T) {
	t.Run("successful_response_with_accrual", func(t *testing.T) {
		expectedAccrual := 100.5
		expectedOrder := "12345678903"
		expectedStatus := "PROCESSED"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/orders/12345678903", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := model.AccrualResponse{
				Order:   expectedOrder,
				Status:  expectedStatus,
				Accrual: &expectedAccrual,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedOrder, result.Order)
		assert.Equal(t, expectedStatus, result.Status)
		assert.NotNil(t, result.Accrual)
		assert.Equal(t, expectedAccrual, *result.Accrual)
	})

	t.Run("successful_response_without_accrual", func(t *testing.T) {
		expectedOrder := "12345678903"
		expectedStatus := "PROCESSING"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := model.AccrualResponse{
				Order:  expectedOrder,
				Status: expectedStatus,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedOrder, result.Order)
		assert.Equal(t, expectedStatus, result.Status)
		assert.Nil(t, result.Accrual)
	})

	t.Run("order_not_found_204", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "99999999999")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "order not found in accrual system")
	})

	t.Run("rate_limited_with_retry_after_header", func(t *testing.T) {
		expectedRetryAfter := 60

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", expectedRetryAfter))
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "rate limited")
		assert.Contains(t, err.Error(), fmt.Sprintf("retry after %d seconds", expectedRetryAfter))
	})

	t.Run("rate_limited_without_retry_after_header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "rate limited")
	})

	t.Run("unexpected_status_code", func(t *testing.T) {
		expectedBody := "internal server error"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(expectedBody))
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected status code: 500")
		assert.Contains(t, err.Error(), expectedBody)
	})

	t.Run("invalid_json_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json{"))
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(15 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("server_unreachable", func(t *testing.T) {
		client := NewAccrualClient("http://localhost:99999")
		ctx := context.Background()

		result, err := client.GetOrderInfo(ctx, "12345678903")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("all_accrual_statuses", func(t *testing.T) {
		statuses := []string{"REGISTERED", "INVALID", "PROCESSING", "PROCESSED"}

		for _, status := range statuses {
			t.Run(fmt.Sprintf("status_%s", status), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)

					response := model.AccrualResponse{
						Order:  "12345678903",
						Status: status,
					}
					json.NewEncoder(w).Encode(response)
				}))
				defer server.Close()

				client := NewAccrualClient(server.URL)
				ctx := context.Background()

				result, err := client.GetOrderInfo(ctx, "12345678903")

				require.NoError(t, err)
				assert.Equal(t, status, result.Status)
			})
		}
	})
}
