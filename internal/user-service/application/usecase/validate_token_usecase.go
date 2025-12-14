package usecase

import (
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/service"
)

type ValidateTokenUseCase struct {
	jwtService service.JWTService
}

func NewValidateTokenUseCase(jwtService service.JWTService) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		jwtService: jwtService,
	}
}

type ValidateTokenRequest struct {
	Token string
}

type ValidateTokenResponse struct {
	Claims *model.Claims
}

func (uc *ValidateTokenUseCase) Execute(req ValidateTokenRequest) (*ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, errors.ErrTokenRequired
	}

	claims, err := uc.jwtService.ValidateToken(req.Token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	return &ValidateTokenResponse{Claims: claims}, nil
}

