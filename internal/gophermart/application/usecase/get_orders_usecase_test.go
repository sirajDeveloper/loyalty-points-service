package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

func TestGetOrdersUseCase_Execute(t *testing.T) {
	now := time.Now()
	accrual1 := 100.5
	accrual2 := 200.0

	tests := []struct {
		name      string
		userID    int64
		setupMock func(*MockOrderRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successful get orders",
			userID: 1,
			setupMock: func(m *MockOrderRepository) {
				orders := []*model.Order{
					model.RestoreOrder(1, 1, "12345678903", model.OrderStatusProcessed, &accrual1, now),
					model.RestoreOrder(2, 1, "79927398713", model.OrderStatusProcessing, nil, now.Add(time.Hour)),
				}
				m.On("FindByUserID", mock.Anything, int64(1)).Return(orders, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "empty orders list",
			userID: 2,
			setupMock: func(m *MockOrderRepository) {
				m.On("FindByUserID", mock.Anything, int64(2)).Return([]*model.Order{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "repository error",
			userID: 3,
			setupMock: func(m *MockOrderRepository) {
				m.On("FindByUserID", mock.Anything, int64(3)).Return(nil, errors.New("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:   "single order with accrual",
			userID: 4,
			setupMock: func(m *MockOrderRepository) {
				orders := []*model.Order{
					model.RestoreOrder(1, 4, "4532015112830366", model.OrderStatusProcessed, &accrual2, now),
				}
				m.On("FindByUserID", mock.Anything, int64(4)).Return(orders, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			tt.setupMock(mockRepo)

			uc := NewGetOrdersUseCase(mockRepo)
			resp, err := uc.Execute(context.Background(), GetOrdersRequest{UserID: tt.userID})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp, tt.wantCount)

				if tt.wantCount > 0 {
					assert.NotEmpty(t, resp[0].Number)
					assert.NotEmpty(t, resp[0].Status)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
