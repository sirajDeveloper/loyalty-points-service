package handler

import (
	"encoding/json"
	"net/http"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
)

type LoginHandler struct {
	loginUseCase *usecase.LoginUseCase
}

func NewLoginHandler(loginUseCase *usecase.LoginUseCase) *LoginHandler {
	return &LoginHandler{
		loginUseCase: loginUseCase,
	}
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	resp, err := h.loginUseCase.Execute(r.Context(), usecase.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, errors.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if errors.Is(err, errors.ErrLoginRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+resp.Token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": resp.Token})
}

