package bootstrap

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	gophermartusecase "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
)

type App struct {
	config          *Config
	server          *http.Server
	processOrdersUC *gophermartusecase.ProcessOrdersUseCase
	pool            *pgxpool.Pool
	workerCtx       context.Context
	workerCancel    context.CancelFunc
}

func NewApp(cfg *Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Initialize() error {
	infrastructureInitializer := NewInfrastructureInitializer(a.config)
	infraResult, err := infrastructureInitializer.Initialize()
	if err != nil {
		return err
	}

	a.pool = infraResult.Pool

	useCaseInitializer := NewUseCaseInitializer(a.config, infraResult)
	useCaseResult := useCaseInitializer.Initialize()

	handlerInitializer := NewHandlerInitializer(a.config, useCaseResult)
	handlerResult := handlerInitializer.Initialize()

	a.server = handlerResult.Server
	a.processOrdersUC = useCaseResult.ProcessOrdersUseCase

	return nil
}

func (a *App) Run() error {
	a.workerCtx, a.workerCancel = context.WithCancel(context.Background())
	go a.processOrdersUC.StartWorker(a.workerCtx, 5*time.Second)

	go func() {
		log.Printf("gophermart service starting on %s", a.config.RunAddress)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("shutting down server...")

	if a.workerCancel != nil {
		a.workerCancel()
	}

	if a.pool != nil {
		defer a.pool.Close()
	}

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("server exited")
	return nil
}
