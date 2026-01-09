package bootstrap

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	gophermartrepository "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/domain/repository"
	gophermartpostgres "github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/datastorage/postgres"
	userservicebootstrap "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/bootstrap"
	userservicerepository "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/repository"
	userserviceservice "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/service"
	userservicepostgres "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/datastorage/postgres"
	userservicejwt "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/infrastructure/jwt"
)

type InfrastructureInitializer struct {
	config *Config
}

func NewInfrastructureInitializer(cfg *Config) *InfrastructureInitializer {
	return &InfrastructureInitializer{
		config: cfg,
	}
}

type InfrastructureResult struct {
	Pool           *pgxpool.Pool
	UserRepo       userservicerepository.UserRepository
	JWTService     userserviceservice.JWTService
	OrderRepo      gophermartrepository.OrderRepository
	BalanceRepo    gophermartrepository.BalanceRepository
	WithdrawalRepo gophermartrepository.WithdrawalRepository
	OutboxRepo     gophermartrepository.OutboxRepository
	UnitOfWork     gophermartrepository.UnitOfWork
	UserServiceCfg *userservicebootstrap.Config
}

func (i *InfrastructureInitializer) Initialize() (*InfrastructureResult, error) {
	if i.config.DatabaseURI == "" {
		log.Fatal("DATABASE_URI is required")
	}

	pool, err := pgxpool.New(context.Background(), i.config.DatabaseURI)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	userServiceCfg := userservicebootstrap.ConfigLoad()
	userServiceCfg.DatabaseURI = i.config.DatabaseURI
	userServiceCfg.JWTSecret = i.config.JWTSecret
	userServiceCfg.JWTExpiry = i.config.JWTExpiry

	userRepo := userservicepostgres.NewUserRepository(pool)
	jwtService := userservicejwt.NewJWTService(userServiceCfg.JWTSecret, userServiceCfg.JWTExpiry)

	orderRepo := gophermartpostgres.NewOrderRepository(pool)
	balanceRepo := gophermartpostgres.NewBalanceRepository(pool)
	withdrawalRepo := gophermartpostgres.NewWithdrawalRepository(pool)
	outboxRepo := gophermartpostgres.NewOutboxRepository(pool)
	unitOfWork := gophermartpostgres.NewUnitOfWork(pool)

	return &InfrastructureResult{
		Pool:           pool,
		UserRepo:       userRepo,
		JWTService:     jwtService,
		OrderRepo:      orderRepo,
		BalanceRepo:    balanceRepo,
		WithdrawalRepo: withdrawalRepo,
		OutboxRepo:     outboxRepo,
		UnitOfWork:     unitOfWork,
		UserServiceCfg: userServiceCfg,
	}, nil
}
