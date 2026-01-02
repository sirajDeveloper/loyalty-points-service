package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/infrastructure/httpclient"
)

type AuthMiddleware struct {
	userServiceClient *httpclient.UserServiceClient
}

func NewAuthMiddleware(userServiceClient *httpclient.UserServiceClient) *AuthMiddleware {
	return &AuthMiddleware{
		userServiceClient: userServiceClient,
	}
}

type contextKey string

const userIDKey contextKey = "userID"

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		claims, err := m.userServiceClient.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

