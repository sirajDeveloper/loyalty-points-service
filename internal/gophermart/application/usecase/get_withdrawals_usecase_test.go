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

func TestGetWithdrawalsUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userID    int64
		setupMock func(*MockWithdrawalRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successful get withdrawals",
			userID: 1,
			setupMock: func(m *MockWithdrawalRepository) {
				withdrawals := []*model.Withdrawal{
					model.RestoreWithdrawal(1, 1, "12345678903", 100.0, now),
					model.RestoreWithdrawal(2, 1, "79927398713", 50.5, now.Add(time.Hour)),
				}
				m.On("FindByUserID", mock.Anything, int64(1)).Return(withdrawals, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "empty withdrawals list",
			userID: 2,
			setupMock: func(m *MockWithdrawalRepository) {
				m.On("FindByUserID", mock.Anything, int64(2)).Return([]*model.Withdrawal{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "repository error",
			userID: 3,
			setupMock: func(m *MockWithdrawalRepository) {
				m.On("FindByUserID", mock.Anything, int64(3)).Return(nil, errors.New("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:   "single withdrawal",
			userID: 4,
			setupMock: func(m *MockWithdrawalRepository) {
				withdrawals := []*model.Withdrawal{
					model.RestoreWithdrawal(1, 4, "4532015112830366", 250.75, now),
				}
				m.On("FindByUserID", mock.Anything, int64(4)).Return(withdrawals, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWithdrawalRepository)
			tt.setupMock(mockRepo)

			uc := NewGetWithdrawalsUseCase(mockRepo)
			resp, err := uc.Execute(context.Background(), GetWithdrawalsRequest{UserID: tt.userID})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp, tt.wantCount)

				if tt.wantCount > 0 {
					assert.NotEmpty(t, resp[0].Order)
					assert.Greater(t, resp[0].Sum, 0.0)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
