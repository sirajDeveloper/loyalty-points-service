package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

func TestGetBalanceUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		setupMock     func(*MockBalanceRepository)
		wantCurrent   float64
		wantWithdrawn float64
		wantErr       bool
	}{
		{
			name:   "successful get balance",
			userID: 1,
			setupMock: func(m *MockBalanceRepository) {
				balance := model.RestoreBalance(1, 100.5, 50.0)
				m.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
			},
			wantCurrent:   100.5,
			wantWithdrawn: 50.0,
			wantErr:       false,
		},
		{
			name:   "new user with zero balance",
			userID: 2,
			setupMock: func(m *MockBalanceRepository) {
				balance := model.NewBalance(2)
				m.On("GetByUserID", mock.Anything, int64(2)).Return(balance, nil)
			},
			wantCurrent:   0.0,
			wantWithdrawn: 0.0,
			wantErr:       false,
		},
		{
			name:   "repository error",
			userID: 3,
			setupMock: func(m *MockBalanceRepository) {
				m.On("GetByUserID", mock.Anything, int64(3)).Return(nil, errors.New("database error"))
			},
			wantCurrent:   0.0,
			wantWithdrawn: 0.0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockBalanceRepository)
			tt.setupMock(mockRepo)

			uc := NewGetBalanceUseCase(mockRepo)
			resp, err := uc.Execute(context.Background(), GetBalanceRequest{UserID: tt.userID})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantCurrent, resp.Current)
				assert.Equal(t, tt.wantWithdrawn, resp.Withdrawn)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
