package bootstrap

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	gophermarthttpclient "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/httpclient"
	gophermarthandler "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/handler"
	gophermartmiddleware "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/middleware"
	userservicehandler "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/presentation/handler"
)

type HandlerInitializer struct {
	config        *Config
	useCaseResult *UseCaseResult
}

func NewHandlerInitializer(cfg *Config, useCaseResult *UseCaseResult) *HandlerInitializer {
	return &HandlerInitializer{
		config:        cfg,
		useCaseResult: useCaseResult,
	}
}

type HandlerResult struct {
	Server *http.Server
}

func (h *HandlerInitializer) Initialize() *HandlerResult {
	registerHandler := userservicehandler.NewRegisterHandler(h.useCaseResult.RegisterUseCase)
	loginHandler := userservicehandler.NewLoginHandler(h.useCaseResult.LoginUseCase)
	validateHandler := userservicehandler.NewValidateHandler(h.useCaseResult.ValidateUseCase)
	healthHandler := userservicehandler.NewHealthHandler()

	orderHandler := gophermarthandler.NewOrderHandler(h.useCaseResult.UploadOrderUseCase, h.useCaseResult.GetOrdersUseCase)
	balanceHandler := gophermarthandler.NewBalanceHandler(h.useCaseResult.GetBalanceUseCase, h.useCaseResult.WithdrawUseCase)
	withdrawalHandler := gophermarthandler.NewWithdrawalHandler(h.useCaseResult.GetWithdrawalsUseCase)

	userServiceClient := gophermarthttpclient.NewUserServiceClient("http://" + h.config.RunAddress)
	authMiddleware := gophermartmiddleware.NewAuthMiddleware(userServiceClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Post("/api/user/register", registerHandler.ServeHTTP)
	r.Post("/api/user/login", loginHandler.ServeHTTP)
	r.Post("/api/auth/validate", validateHandler.ServeHTTP)
	r.Get("/api/auth/health", healthHandler.ServeHTTP)

	r.With(authMiddleware.Handle).Post("/api/user/orders", orderHandler.Upload)
	r.With(authMiddleware.Handle).Get("/api/user/orders", orderHandler.GetList)
	r.With(authMiddleware.Handle).Get("/api/user/balance", balanceHandler.Get)
	r.With(authMiddleware.Handle).Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
	r.With(authMiddleware.Handle).Get("/api/user/withdrawals", withdrawalHandler.GetList)

	server := &http.Server{
		Addr:    h.config.RunAddress,
		Handler: r,
	}

	return &HandlerResult{
		Server: server,
	}
}
