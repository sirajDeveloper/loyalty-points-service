package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID int64, login string) (string, error) {
	args := m.Called(userID, login)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*model.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Claims), args.Error(1)
}
