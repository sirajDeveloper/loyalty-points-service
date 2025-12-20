package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/middleware"
)

type BalanceHandler struct {
	getBalanceUseCase *usecase.GetBalanceUseCase
	withdrawUseCase   *usecase.WithdrawUseCase
}

func NewBalanceHandler(
	getBalanceUseCase *usecase.GetBalanceUseCase,
	withdrawUseCase *usecase.WithdrawUseCase,
) *BalanceHandler {
	return &BalanceHandler{
		getBalanceUseCase: getBalanceUseCase,
		withdrawUseCase:   withdrawUseCase,
	}
}

func (h *BalanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.getBalanceUseCase.Execute(r.Context(), usecase.GetBalanceRequest{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	_, err := h.withdrawUseCase.Execute(r.Context(), usecase.WithdrawRequest{
		UserID: userID,
		Order:  req.Order,
		Sum:    req.Sum,
	})
	if err != nil {
		if err.Error() == "invalid order number format" {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if err.Error() == "insufficient funds" {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}
		log.Printf("withdraw error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
