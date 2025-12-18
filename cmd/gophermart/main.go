package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	userserviceusecase "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
	userservicehandler "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/presentation/handler"
	userservicebootstrap "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/bootstrap"

	gophermartusecase "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
	gophermarthandler "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/handler"
	gophermartmiddleware "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/middleware"
	gophermartbootstrap "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/bootstrap"
	gophermartpostgres "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/datastorage/postgres"
	gophermarthttpclient "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/httpclient"
	gophermartservice "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/service"

	userservicepostgres "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/datastorage/postgres"
	userservicejwt "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/jwt"
)

func main() {
	cfg := gophermartbootstrap.ConfigLoad()

	if cfg.DatabaseURI == "" {
		log.Fatal("DATABASE_URI is required")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	userServiceCfg := userservicebootstrap.ConfigLoad()
	userServiceCfg.DatabaseURI = cfg.DatabaseURI
	userServiceCfg.JWTSecret = cfg.JWTSecret
	userServiceCfg.JWTExpiry = cfg.JWTExpiry

	userRepo := userservicepostgres.NewUserRepository(pool)
	jwtService := userservicejwt.NewJWTService(userServiceCfg.JWTSecret, userServiceCfg.JWTExpiry)

	registerUseCase := userserviceusecase.NewRegisterUseCase(userRepo, jwtService)
	loginUseCase := userserviceusecase.NewLoginUseCase(userRepo, jwtService)
	validateUseCase := userserviceusecase.NewValidateTokenUseCase(jwtService)

	registerHandler := userservicehandler.NewRegisterHandler(registerUseCase)
	loginHandler := userservicehandler.NewLoginHandler(loginUseCase)
	validateHandler := userservicehandler.NewValidateHandler(validateUseCase)
	healthHandler := userservicehandler.NewHealthHandler()

	orderRepo := gophermartpostgres.NewOrderRepository(pool)
	balanceRepo := gophermartpostgres.NewBalanceRepository(pool)
	withdrawalRepo := gophermartpostgres.NewWithdrawalRepository(pool)
	outboxRepo := gophermartpostgres.NewOutboxRepository(pool)
	accrualClient := gophermarthttpclient.NewAccrualClient(cfg.AccrualSystemAddress)
	luhnValidator := gophermartservice.NewLuhnValidator()

	uploadOrderUseCase := gophermartusecase.NewUploadOrderUseCase(pool, orderRepo, outboxRepo, luhnValidator)
	getOrdersUseCase := gophermartusecase.NewGetOrdersUseCase(orderRepo)
	getBalanceUseCase := gophermartusecase.NewGetBalanceUseCase(balanceRepo)
	withdrawUseCase := gophermartusecase.NewWithdrawUseCase(pool, balanceRepo, withdrawalRepo, luhnValidator)
	getWithdrawalsUseCase := gophermartusecase.NewGetWithdrawalsUseCase(withdrawalRepo)
	processOrdersUseCase := gophermartusecase.NewProcessOrdersUseCase(outboxRepo, orderRepo, balanceRepo, accrualClient)

	orderHandler := gophermarthandler.NewOrderHandler(uploadOrderUseCase, getOrdersUseCase)
	balanceHandler := gophermarthandler.NewBalanceHandler(getBalanceUseCase, withdrawUseCase)
	withdrawalHandler := gophermarthandler.NewWithdrawalHandler(getWithdrawalsUseCase)

	userServiceClient := gophermarthttpclient.NewUserServiceClient("http://" + cfg.RunAddress)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go processOrdersUseCase.StartWorker(ctx, 5*time.Second)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	go func() {
		log.Printf("gophermart service starting on %s", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
