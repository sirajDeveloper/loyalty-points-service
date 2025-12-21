package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainerrors "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		req        RegisterRequest
		setupMocks func(*MockUserRepository, *MockJWTService)
		want       *RegisterResponse
		wantErr    bool
		errType    error
	}{
		{
			name: "successful registration",
			req: RegisterRequest{
				Login:     "testuser",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("ExistsByLogin", mock.Anything, "testuser").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.Login == "testuser" && u.FirstName == "Test" && u.LastName == "User" && u.PasswordHash != ""
				})).Return(nil)
				jwtService.On("GenerateToken", mock.Anything, "testuser").Return("test-token", nil)
			},
			want:    &RegisterResponse{Token: "test-token"},
			wantErr: false,
		},
		{
			name: "empty login",
			req: RegisterRequest{
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
			req: RegisterRequest{
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
			name: "login already exists",
			req: RegisterRequest{
				Login:    "existinguser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("ExistsByLogin", mock.Anything, "existinguser").Return(true, nil)
			},
			want:    nil,
			wantErr: true,
			errType: domainerrors.ErrLoginAlreadyExists,
		},
		{
			name: "repository exists check error",
			req: RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("ExistsByLogin", mock.Anything, "testuser").Return(false, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "repository create error",
			req: RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("ExistsByLogin", mock.Anything, "testuser").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("create error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "jwt generation error",
			req: RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, jwtService *MockJWTService) {
				userRepo.On("ExistsByLogin", mock.Anything, "testuser").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				jwtService.On("GenerateToken", mock.Anything, "testuser").Return("", errors.New("jwt error"))
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

			uc := NewRegisterUseCase(userRepo, jwtService)
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
