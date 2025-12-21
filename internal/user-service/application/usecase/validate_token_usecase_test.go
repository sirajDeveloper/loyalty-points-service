package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/mock"

	domainerrors "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
)

func TestValidateTokenUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		req        ValidateTokenRequest
		setupMocks func(*MockJWTService)
		want       *ValidateTokenResponse
		wantErr    bool
		errType    error
	}{
		{
			name: "successful validation",
			req: ValidateTokenRequest{
				Token: "valid-token",
			},
			setupMocks: func(jwtService *MockJWTService) {
				claims := &model.Claims{
					UserID: 1,
					Login:  "testuser",
				}
				jwtService.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			want: &ValidateTokenResponse{
				Claims: &model.Claims{
					UserID: 1,
					Login:  "testuser",
				},
			},
			wantErr: false,
		},
		{
			name: "empty token",
			req: ValidateTokenRequest{
				Token: "",
			},
			setupMocks: func(jwtService *MockJWTService) {
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrTokenRequired,
		},
		{
			name: "invalid token",
			req: ValidateTokenRequest{
				Token: "invalid-token",
			},
			setupMocks: func(jwtService *MockJWTService) {
				jwtService.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrInvalidToken,
		},
		{
			name: "jwt service error",
			req: ValidateTokenRequest{
				Token: "token",
			},
			setupMocks: func(jwtService *MockJWTService) {
				jwtService.On("ValidateToken", "token").Return(nil, errors.New("jwt error"))
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := new(MockJWTService)
			tt.setupMocks(jwtService)

			uc := NewValidateTokenUseCase(jwtService)
			got, err := uc.Execute(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			jwtService.AssertExpectations(t)
		})
	}
}
