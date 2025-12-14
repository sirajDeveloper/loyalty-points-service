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
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/bootstrap"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/datastorage/postgres"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/jwt"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/presentation/handler"
)

func main() {
	cfg := bootstrap.ConfigLoad()

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

	userRepo := postgres.NewUserRepository(pool)
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	registerUseCase := usecase.NewRegisterUseCase(userRepo, jwtService)
	loginUseCase := usecase.NewLoginUseCase(userRepo, jwtService)
	validateUseCase := usecase.NewValidateTokenUseCase(jwtService)

	registerHandler := handler.NewRegisterHandler(registerUseCase)
	loginHandler := handler.NewLoginHandler(loginUseCase)
	validateHandler := handler.NewValidateHandler(validateUseCase)
	healthHandler := handler.NewHealthHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Post("/api/user/register", registerHandler.ServeHTTP)
	r.Post("/api/user/login", loginHandler.ServeHTTP)
	r.Post("/api/auth/validate", validateHandler.ServeHTTP)
	r.Get("/api/auth/health", healthHandler.ServeHTTP)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	go func() {
		log.Printf("user-service starting on %s", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
