package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
)

type MockOrderNumberValidator struct {
	mock.Mock
}

func (m *MockOrderNumberValidator) Validate(number string) bool {
	args := m.Called(number)
	return args.Bool(0)
}

type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Begin(ctx context.Context) (repository.Transaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Transaction), args.Error(1)
}

type MockTransaction struct {
	mock.Mock
	balanceRepo    *MockBalanceRepository
	withdrawalRepo *MockWithdrawalRepository
	orderRepo      *MockOrderRepository
	outboxRepo     *MockOutboxRepository
}

func (m *MockTransaction) BalanceRepository() repository.BalanceRepository {
	return m.balanceRepo
}

func (m *MockTransaction) WithdrawalRepository() repository.WithdrawalRepository {
	return m.withdrawalRepo
}

func (m *MockTransaction) OrderRepository() repository.OrderRepository {
	return m.orderRepo
}

func (m *MockTransaction) OutboxRepository() repository.OutboxRepository {
	return m.outboxRepo
}

func (m *MockTransaction) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransaction) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) GetByUserID(ctx context.Context, userID int64) (*model.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *MockBalanceRepository) Withdraw(ctx context.Context, userID int64, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockBalanceRepository) Accrue(ctx context.Context, userID int64, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

type MockWithdrawalRepository struct {
	mock.Mock
}

func (m *MockWithdrawalRepository) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	args := m.Called(ctx, withdrawal)
	return args.Error(0)
}

func (m *MockWithdrawalRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Withdrawal), args.Error(1)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByNumber(ctx context.Context, number string) (*model.Order, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status model.OrderStatus, accrual *float64) error {
	args := m.Called(ctx, orderID, status, accrual)
	return args.Error(0)
}

func (m *MockOrderRepository) FindPending(ctx context.Context, limit int) ([]*model.Order, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) Create(ctx context.Context, outbox *model.Outbox) error {
	args := m.Called(ctx, outbox)
	return args.Error(0)
}

func (m *MockOutboxRepository) FindPending(ctx context.Context, limit int) ([]*model.Outbox, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Outbox), args.Error(1)
}

func (m *MockOutboxRepository) UpdateStatus(ctx context.Context, id int64, status model.OutboxStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOutboxRepository) IncrementRetries(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
