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

type MockAccrualService struct {
	mock.Mock
}

func (m *MockAccrualService) GetOrderInfo(ctx context.Context, orderNumber string) (*model.AccrualResponse, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccrualResponse), args.Error(1)
}

func TestProcessOrdersUseCase_ProcessPendingOrders(t *testing.T) {
	now := time.Now()
	accrual := 100.5

	tests := []struct {
		name         string
		setupOutbox  func(*MockOutboxRepository)
		setupOrder   func(*MockOrderRepository)
		setupBalance func(*MockBalanceRepository)
		setupAccrual func(*MockAccrualService)
		wantErr      bool
	}{
		{
			name: "successful process single order with PROCESSED status",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OutboxStatusProcessed).Return(nil)
			},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusProcessed, &accrual).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
				m.On("Accrue", mock.Anything, int64(1), accrual).Return(nil)
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:   "79927398713",
					Status:  "PROCESSED",
					Accrual: &accrual,
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name: "empty pending orders",
			setupOutbox: func(m *MockOutboxRepository) {
				m.On("FindPending", mock.Anything, 10).Return([]*model.Outbox{}, nil)
			},
			setupOrder: func(m *MockOrderRepository) {
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: false,
		},
		{
			name: "outbox repository find error",
			setupOutbox: func(m *MockOutboxRepository) {
				m.On("FindPending", mock.Anything, 10).Return(nil, errors.New("database error"))
			},
			setupOrder: func(m *MockOrderRepository) {
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: true,
		},
		{
			name: "process order error with retries increment",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 1, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("IncrementRetries", mock.Anything, int64(1)).Return(nil)
			},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("order not found"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: false,
		},
		{
			name: "process order error with max retries - mark as failed",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 3, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OutboxStatusFailed).Return(nil)
			},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("order not found"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: false,
		},
		{
			name: "error updating outbox status to failed",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 3, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OutboxStatusFailed).Return(errors.New("update failed error"))
			},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("order not found"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: false,
		},
		{
			name: "error incrementing retries",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 1, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("IncrementRetries", mock.Anything, int64(1)).Return(errors.New("increment error"))
			},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("order not found"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: false,
		},
		{
			name: "error updating outbox status to processed",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OutboxStatusProcessed).Return(errors.New("update processed error"))
			},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusNew, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "REGISTERED",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name: "multiple orders - some succeed, some fail",
			setupOutbox: func(m *MockOutboxRepository) {
				outboxes := []*model.Outbox{
					{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0, CreatedAt: now, UpdatedAt: now},
					{ID: 2, OrderID: 2, Status: model.OutboxStatusPending, Retries: 0, CreatedAt: now, UpdatedAt: now},
				}
				m.On("FindPending", mock.Anything, 10).Return(outboxes, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OutboxStatusProcessed).Return(nil)
				m.On("IncrementRetries", mock.Anything, int64(2)).Return(nil)
			},
			setupOrder: func(m *MockOrderRepository) {
				order1 := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order1, nil).Once()
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusProcessed, &accrual).Return(nil)
				m.On("FindByID", mock.Anything, int64(2)).Return(nil, errors.New("order not found")).Once()
			},
			setupBalance: func(m *MockBalanceRepository) {
				m.On("Accrue", mock.Anything, int64(1), accrual).Return(nil)
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:   "79927398713",
					Status:  "PROCESSED",
					Accrual: &accrual,
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOutboxRepo := new(MockOutboxRepository)
			mockOrderRepo := new(MockOrderRepository)
			mockBalanceRepo := new(MockBalanceRepository)
			mockAccrualService := new(MockAccrualService)

			tt.setupOutbox(mockOutboxRepo)
			tt.setupOrder(mockOrderRepo)
			tt.setupBalance(mockBalanceRepo)
			tt.setupAccrual(mockAccrualService)

			uc := NewProcessOrdersUseCase(mockOutboxRepo, mockOrderRepo, mockBalanceRepo, mockAccrualService)
			err := uc.ProcessPendingOrders(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockOutboxRepo.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
			mockBalanceRepo.AssertExpectations(t)
			mockAccrualService.AssertExpectations(t)
		})
	}
}

func TestProcessOrdersUseCase_processOrder(t *testing.T) {
	now := time.Now()
	accrual := 150.75

	tests := []struct {
		name         string
		outbox       *model.Outbox
		setupOrder   func(*MockOrderRepository)
		setupBalance func(*MockBalanceRepository)
		setupAccrual func(*MockAccrualService)
		wantErr      bool
		errMsg       string
	}{
		{
			name:   "successful process REGISTERED status",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusNew, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "REGISTERED",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name:   "successful process PROCESSING status",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusProcessing, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "PROCESSING",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name:   "successful process INVALID status",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusInvalid, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "INVALID",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name:   "successful process PROCESSED status with accrual",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessing, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusProcessed, &accrual).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
				m.On("Accrue", mock.Anything, int64(1), accrual).Return(nil)
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:   "79927398713",
					Status:  "PROCESSED",
					Accrual: &accrual,
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name:   "PROCESSED status without accrual",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessing, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusProcessed, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "PROCESSED",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: false,
		},
		{
			name:   "order not found",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: true,
			errMsg:  "order not found",
		},
		{
			name:   "order repository find error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				m.On("FindByID", mock.Anything, int64(1)).Return(nil, errors.New("database error"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
			},
			wantErr: true,
		},
		{
			name:   "order not found in accrual system",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusInvalid, (*float64)(nil)).Return(nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(nil, errors.New("order not found in accrual system"))
			},
			wantErr: false,
		},
		{
			name:   "order not found in accrual system - order update status error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessed, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(nil, errors.New("order not found in accrual system"))
			},
			wantErr: true,
		},
		{
			name:   "order not found in accrual system - repository update error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusInvalid, (*float64)(nil)).Return(errors.New("repository error"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(nil, errors.New("order not found in accrual system"))
			},
			wantErr: true,
		},
		{
			name:   "rate limited error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(nil, errors.New("rate limited"))
			},
			wantErr: true,
			errMsg:  "rate limited",
		},
		{
			name:   "unknown accrual status",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "UNKNOWN",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
			errMsg:  "unknown accrual status",
		},
		{
			name:   "balance accrue error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessing, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
				m.On("Accrue", mock.Anything, int64(1), accrual).Return(errors.New("balance error"))
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:   "79927398713",
					Status:  "PROCESSED",
					Accrual: &accrual,
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
		},
		{
			name:   "order UpdateStatus error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				// Order with status PROCESSED cannot be updated to NEW
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessed, &accrual, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "NEW",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
		},
		{
			name:   "order repository UpdateStatus error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusNew, (*float64)(nil)).Return(errors.New("repository update error"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "REGISTERED",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
		},
		{
			name:   "accrual service generic error (not rate limited, not order not found)",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(nil, errors.New("accrual service internal error"))
			},
			wantErr: true,
			errMsg:  "accrual service internal error",
		},
		{
			name:   "order update status error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessed, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "PROCESSING",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
		},
		{
			name:   "order repository update status error",
			outbox: &model.Outbox{ID: 1, OrderID: 1, Status: model.OutboxStatusPending, Retries: 0},
			setupOrder: func(m *MockOrderRepository) {
				order := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, now)
				m.On("FindByID", mock.Anything, int64(1)).Return(order, nil)
				m.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusNew, (*float64)(nil)).Return(errors.New("update error"))
			},
			setupBalance: func(m *MockBalanceRepository) {
			},
			setupAccrual: func(m *MockAccrualService) {
				resp := &model.AccrualResponse{
					Order:  "79927398713",
					Status: "REGISTERED",
				}
				m.On("GetOrderInfo", mock.Anything, "79927398713").Return(resp, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOutboxRepo := new(MockOutboxRepository)
			mockOrderRepo := new(MockOrderRepository)
			mockBalanceRepo := new(MockBalanceRepository)
			mockAccrualService := new(MockAccrualService)

			tt.setupOrder(mockOrderRepo)
			tt.setupBalance(mockBalanceRepo)
			tt.setupAccrual(mockAccrualService)

			uc := NewProcessOrdersUseCase(mockOutboxRepo, mockOrderRepo, mockBalanceRepo, mockAccrualService)
			err := uc.processOrder(context.Background(), tt.outbox)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockOrderRepo.AssertExpectations(t)
			mockBalanceRepo.AssertExpectations(t)
			mockAccrualService.AssertExpectations(t)
		})
	}
}
