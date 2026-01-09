package usecase

import (
	"context"

	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/repository"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/service"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	userRepo   repository.UserRepository
	jwtService service.JWTService
}

func NewLoginUseCase(userRepo repository.UserRepository, jwtService service.JWTService) *LoginUseCase {
	return &LoginUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type LoginRequest struct {
	Login    string
	Password string
}

type LoginResponse struct {
	Token string
}

func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Login == "" || req.Password == "" {
		return nil, errors.ErrLoginRequired
	}

	user, err := uc.userRepo.FindByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}
