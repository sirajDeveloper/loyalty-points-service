package service

import "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"

type JWTService interface {
	GenerateToken(userID int64, login string) (string, error)
	ValidateToken(tokenString string) (*model.Claims, error)
}


