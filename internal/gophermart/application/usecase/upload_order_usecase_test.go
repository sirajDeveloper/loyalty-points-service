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

func TestUploadOrderUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		req            UploadOrderRequest
		setupValidator func(*MockOrderNumberValidator)
		setupOrderRepo func(*MockOrderRepository)
		setupUOW       func(*MockUnitOfWork, *MockTransaction, *MockOrderRepository, *MockOutboxRepository)
		wantStatus     string
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful upload new order",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				orderRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.UserID() == 1 && o.Number() == "79927398713"
				})).Return(nil)
				outboxRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *model.Outbox) bool {
					return o.Status == model.OutboxStatusPending && o.Retries == 0
				})).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantStatus: "accepted",
			wantErr:    false,
		},
		{
			name: "invalid order number format",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "invalid",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "invalid").Return(false)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
			},
			wantStatus: "",
			wantErr:    true,
			errMsg:     "invalid order number format",
		},
		{
			name: "order already uploaded by same user",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				existingOrder := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusNew, nil, time.Now())
				m.On("FindByNumber", mock.Anything, "79927398713").Return(existingOrder, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
			},
			wantStatus: "already_uploaded",
			wantErr:    false,
		},
		{
			name: "order already exists for another user",
			req: UploadOrderRequest{
				UserID: 2,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				existingOrder := model.RestoreOrder(1, 1, "79927398713", model.OrderStatusProcessing, nil, time.Now())
				m.On("FindByNumber", mock.Anything, "79927398713").Return(existingOrder, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
			},
			wantStatus: "",
			wantErr:    true,
			errMsg:     "order number already exists for another user",
		},
		{
			name: "order repository find error",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, errors.New("database error"))
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
			},
			wantStatus: "",
			wantErr:    true,
		},
		{
			name: "unit of work begin error",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				uow.On("Begin", mock.Anything).Return(nil, errors.New("database connection error"))
			},
			wantStatus: "",
			wantErr:    true,
		},
		{
			name: "order create error",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				orderRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantStatus: "",
			wantErr:    true,
		},
		{
			name: "outbox create error",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				orderRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				outboxRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantStatus: "",
			wantErr:    true,
		},
		{
			name: "commit error",
			req: UploadOrderRequest{
				UserID: 1,
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				orderRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				outboxRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(errors.New("commit error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantStatus: "",
			wantErr:    true,
		},
		{
			name: "NewOrder validation error - invalid user ID",
			req: UploadOrderRequest{
				UserID: 0, // Invalid user ID
				Number: "79927398713",
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupOrderRepo: func(m *MockOrderRepository) {
				m.On("FindByNumber", mock.Anything, "79927398713").Return(nil, nil)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, orderRepo *MockOrderRepository, outboxRepo *MockOutboxRepository) {
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantStatus: "",
			wantErr:    true,
			errMsg:     "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := new(MockOrderNumberValidator)
			mockOrderRepo := new(MockOrderRepository)
			mockUOW := new(MockUnitOfWork)
			mockTx := new(MockTransaction)
			mockTxOrderRepo := new(MockOrderRepository)
			mockOutboxRepo := new(MockOutboxRepository)

			mockTx.orderRepo = mockTxOrderRepo
			mockTx.outboxRepo = mockOutboxRepo

			tt.setupValidator(mockValidator)
			tt.setupOrderRepo(mockOrderRepo)
			tt.setupUOW(mockUOW, mockTx, mockTxOrderRepo, mockOutboxRepo)

			uc := NewUploadOrderUseCase(mockUOW, mockOrderRepo, nil, mockValidator)
			resp, err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantStatus, resp.Status)
			}

			mockValidator.AssertExpectations(t)
			mockOrderRepo.AssertExpectations(t)
			if !tt.wantErr {
				mockUOW.AssertExpectations(t)
			}
		})
	}
}
