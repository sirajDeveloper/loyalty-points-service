package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
)

type ValidateHandler struct {
	validateUseCase *usecase.ValidateTokenUseCase
}

func NewValidateHandler(validateUseCase *usecase.ValidateTokenUseCase) *ValidateHandler {
	return &ValidateHandler{
		validateUseCase: validateUseCase,
	}
}

type ValidateRequest struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
}

func (h *ValidateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var token string

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		var req ValidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}
		token = req.Token
	}

	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	resp, err := h.validateUseCase.Execute(usecase.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateResponse{
		UserID: resp.Claims.UserID,
		Login:  resp.Claims.Login,
	})
}
