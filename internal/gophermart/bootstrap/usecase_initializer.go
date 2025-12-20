package bootstrap

import (
	gophermartusecase "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
	gophermartservice "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"
	gophermarthttpclient "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/httpclient"
	userserviceusecase "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
)

type UseCaseInitializer struct {
	config      *Config
	infraResult *InfrastructureResult
}

func NewUseCaseInitializer(cfg *Config, infraResult *InfrastructureResult) *UseCaseInitializer {
	return &UseCaseInitializer{
		config:      cfg,
		infraResult: infraResult,
	}
}

type UseCaseResult struct {
	RegisterUseCase       *userserviceusecase.RegisterUseCase
	LoginUseCase          *userserviceusecase.LoginUseCase
	ValidateUseCase       *userserviceusecase.ValidateTokenUseCase
	UploadOrderUseCase    *gophermartusecase.UploadOrderUseCase
	GetOrdersUseCase      *gophermartusecase.GetOrdersUseCase
	GetBalanceUseCase     *gophermartusecase.GetBalanceUseCase
	WithdrawUseCase       *gophermartusecase.WithdrawUseCase
	GetWithdrawalsUseCase *gophermartusecase.GetWithdrawalsUseCase
	ProcessOrdersUseCase  *gophermartusecase.ProcessOrdersUseCase
}

func (u *UseCaseInitializer) Initialize() *UseCaseResult {
	registerUseCase := userserviceusecase.NewRegisterUseCase(
		u.infraResult.UserRepo,
		u.infraResult.JWTService,
	)
	loginUseCase := userserviceusecase.NewLoginUseCase(
		u.infraResult.UserRepo,
		u.infraResult.JWTService,
	)
	validateUseCase := userserviceusecase.NewValidateTokenUseCase(
		u.infraResult.JWTService,
	)

	accrualClient := gophermarthttpclient.NewAccrualClient(u.config.AccrualSystemAddress)
	luhnValidator := gophermartservice.NewLuhnValidator()

	uploadOrderUseCase := gophermartusecase.NewUploadOrderUseCase(u.infraResult.UnitOfWork, u.infraResult.OrderRepo, u.infraResult.OutboxRepo, luhnValidator)
	getOrdersUseCase := gophermartusecase.NewGetOrdersUseCase(u.infraResult.OrderRepo)
	getBalanceUseCase := gophermartusecase.NewGetBalanceUseCase(u.infraResult.BalanceRepo)
	withdrawUseCase := gophermartusecase.NewWithdrawUseCase(u.infraResult.UnitOfWork, u.infraResult.BalanceRepo, u.infraResult.WithdrawalRepo, luhnValidator)
	getWithdrawalsUseCase := gophermartusecase.NewGetWithdrawalsUseCase(u.infraResult.WithdrawalRepo)
	processOrdersUseCase := gophermartusecase.NewProcessOrdersUseCase(u.infraResult.OutboxRepo, u.infraResult.OrderRepo, u.infraResult.BalanceRepo, accrualClient)

	return &UseCaseResult{
		RegisterUseCase:       registerUseCase,
		LoginUseCase:          loginUseCase,
		ValidateUseCase:       validateUseCase,
		UploadOrderUseCase:    uploadOrderUseCase,
		GetOrdersUseCase:      getOrdersUseCase,
		GetBalanceUseCase:     getBalanceUseCase,
		WithdrawUseCase:       withdrawUseCase,
		GetWithdrawalsUseCase: getWithdrawalsUseCase,
		ProcessOrdersUseCase:  processOrdersUseCase,
	}
}
