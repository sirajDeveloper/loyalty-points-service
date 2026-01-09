package handler

import (
	"encoding/json"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"log"
	"net/http"
)

type RegisterHandler struct {
	registerUseCase *usecase.RegisterUseCase
}

func NewRegisterHandler(registerUseCase *usecase.RegisterUseCase) *RegisterHandler {
	return &RegisterHandler{
		registerUseCase: registerUseCase,
	}
}

type RegisterRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	resp, err := h.registerUseCase.Execute(r.Context(), usecase.RegisterRequest{
		Login:     req.Login,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if errors.Is(err, errors.ErrLoginAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, errors.ErrLoginRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("register error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+resp.Token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"token": resp.Token}); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
