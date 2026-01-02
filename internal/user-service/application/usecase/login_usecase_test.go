package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	domainerrors "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
)

func TestLoginUseCase_Execute(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name       string
		req        LoginRequest
		setupMocks func(*MockUserRepository, *MockJWTService)
		want       *LoginResponse
		wantErr    bool
		errType    error
	}{
		{
			name: "successful login",
			req: LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				user := &model.User{
					ID:           1,
					Login:        "testuser",
					PasswordHash: string(hashedPassword),
				}
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(user, nil)
				jwtService.On("GenerateToken", int64(1), "testuser").Return("test-token", nil)
			},
			want:    &LoginResponse{Token: "test-token"},
			wantErr: false,
		},
		{
			name: "empty login",
			req: LoginRequest{
				Login:    "",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrLoginRequired,
		},
		{
			name: "empty password",
			req: LoginRequest{
				Login:    "testuser",
				Password: "",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrLoginRequired,
		},
		{
			name: "user not found",
			req: LoginRequest{
				Login:    "nonexistent",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("FindByLogin", mock.Anything, "nonexistent").Return(nil, domainerrors.ErrUserNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			req: LoginRequest{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				user := &model.User{
					ID:           1,
					Login:        "testuser",
					PasswordHash: string(hashedPassword),
				}
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(user, nil)
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrInvalidCredentials,
		},
		{
			name: "repository error - not ErrUserNotFound",
			req: LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(nil, errors.New("database connection error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "jwt generation error",
			req: LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				user := &model.User{
					ID:           1,
					Login:        "testuser",
					PasswordHash: string(hashedPassword),
				}
				userRepo.On("FindByLogin", mock.Anything, "testuser").Return(user, nil)
				jwtService.On("GenerateToken", int64(1), "testuser").Return("", errors.New("jwt error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			jwtService := new(MockJWTService)
			tt.setupMocks(userRepo, jwtService)

			uc := NewLoginUseCase(userRepo, jwtService)
			got, err := uc.Execute(context.Background(), tt.req)

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

			userRepo.AssertExpectations(t)
			jwtService.AssertExpectations(t)
		})
	}
}
