package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/model"
)

func TestWithdrawUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		req            WithdrawRequest
		setupValidator func(*MockOrderNumberValidator)
		setupUOW       func(*MockUnitOfWork, *MockTransaction, *MockBalanceRepository, *MockWithdrawalRepository)
		wantSuccess    bool
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful withdrawal",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				balanceRepo.On("Withdraw", mock.Anything, int64(1), 50.0).Return(nil)
				withdrawalRepo.On("Create", mock.Anything, mock.MatchedBy(func(w *model.Withdrawal) bool {
					return w.UserID() == 1 && w.OrderNumber() == "79927398713" && w.Sum() == 50.0
				})).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "invalid order number format",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "invalid",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "invalid").Return(false)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
			},
			wantSuccess: false,
			wantErr:     true,
			errMsg:      "invalid order number format",
		},
		{
			name: "insufficient funds",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    150.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
			errMsg:      "insufficient funds",
		},
		{
			name: "unit of work begin error",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				uow.On("Begin", mock.Anything).Return(nil, errors.New("database connection error"))
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "balance repository get error",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(nil, errors.New("database error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "withdrawal create error",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				withdrawalRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "balance withdraw error",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				balanceRepo.On("Withdraw", mock.Anything, int64(1), 50.0).Return(errors.New("database error"))
				withdrawalRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "commit error",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "79927398713",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "79927398713").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				balanceRepo.On("Withdraw", mock.Anything, int64(1), 50.0).Return(nil)
				withdrawalRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(errors.New("commit error"))
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "NewWithdrawal validation error - invalid user ID",
			req: WithdrawRequest{
				UserID: 0,
				Order:  "12345678903",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "12345678903").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(0, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(0)).Return(balance, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "NewWithdrawal validation error - empty order number",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "",
				Sum:    50.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "").Return(true) // Validator passes, but NewWithdrawal will fail
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "NewWithdrawal validation error - zero sum",
			req: WithdrawRequest{
				UserID: 1,
				Order:  "12345678903",
				Sum:    0.0,
			},
			setupValidator: func(m *MockOrderNumberValidator) {
				m.On("Validate", "12345678903").Return(true)
			},
			setupUOW: func(uow *MockUnitOfWork, tx *MockTransaction, balanceRepo *MockBalanceRepository, withdrawalRepo *MockWithdrawalRepository) {
				balance := model.RestoreBalance(1, 100.0, 0.0)
				balanceRepo.On("GetByUserID", mock.Anything, int64(1)).Return(balance, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				uow.On("Begin", mock.Anything).Return(tx, nil)
			},
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := new(MockOrderNumberValidator)
			mockUOW := new(MockUnitOfWork)
			mockTx := new(MockTransaction)
			mockBalanceRepo := new(MockBalanceRepository)
			mockWithdrawalRepo := new(MockWithdrawalRepository)

			mockTx.balanceRepo = mockBalanceRepo
			mockTx.withdrawalRepo = mockWithdrawalRepo

			tt.setupValidator(mockValidator)
			tt.setupUOW(mockUOW, mockTx, mockBalanceRepo, mockWithdrawalRepo)

			uc := NewWithdrawUseCase(mockUOW, nil, nil, mockValidator)
			resp, err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				if !tt.wantSuccess {
					assert.Nil(t, resp)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantSuccess, resp.Success)
			}

			mockValidator.AssertExpectations(t)
			mockUOW.AssertExpectations(t)
			if tt.wantErr || !tt.wantSuccess {
				mockTx.AssertExpectations(t)
			}
		})
	}
}
