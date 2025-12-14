package model

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID int64
	Login  string
	jwt.RegisteredClaims
}
