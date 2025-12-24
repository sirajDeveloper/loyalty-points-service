package handler

import (
	"encoding/json"
	"net/http"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/middleware"
)

type WithdrawalHandler struct {
	getWithdrawalsUseCase *usecase.GetWithdrawalsUseCase
}

func NewWithdrawalHandler(getWithdrawalsUseCase *usecase.GetWithdrawalsUseCase) *WithdrawalHandler {
	return &WithdrawalHandler{
		getWithdrawalsUseCase: getWithdrawalsUseCase,
	}
}

func (h *WithdrawalHandler) GetList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.getWithdrawalsUseCase.Execute(r.Context(), usecase.GetWithdrawalsRequest{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}


