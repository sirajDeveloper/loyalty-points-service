package usecase

import (
	"context"
	"time"
	"golang.org/x/crypto/bcrypt"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/service"
)

type RegisterUseCase struct {
	userRepo repository.UserRepository
	jwtService service.JWTService
}

func NewRegisterUseCase(userRepo repository.UserRepository, jwtService service.JWTService) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:  userRepo,
		jwtService: jwtService,
	}
}

type RegisterRequest struct {
	Login     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterResponse struct {
	Token string
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	if req.Login == "" || req.Password == "" {
		return nil, errors.ErrLoginRequired
	}

	exists, err := uc.userRepo.ExistsByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrLoginAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		CreatedAt:    time.Now(),
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{Token: token}, nil
}

