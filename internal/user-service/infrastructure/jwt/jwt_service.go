package jwt

import (
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/service"
)

type jwtService struct {
	secretKey []byte
	expiry    time.Duration
}

func NewJWTService(secretKey string, expiry time.Duration) service.JWTService {
	return &jwtService{
		secretKey: []byte(secretKey),
		expiry:    expiry,
	}
}

func (s *jwtService) GenerateToken(userID int64, login string) (string, error) {
	claims := &model.Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *jwtService) ValidateToken(tokenString string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}


