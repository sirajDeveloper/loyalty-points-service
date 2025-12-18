package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/application/usecase"
	"github.com/sirajDeveloper/loyalty-points-service/internal/gophermart/presentation/middleware"
)

type OrderHandler struct {
	uploadOrderUseCase *usecase.UploadOrderUseCase
	getOrdersUseCase   *usecase.GetOrdersUseCase
}

func NewOrderHandler(
	uploadOrderUseCase *usecase.UploadOrderUseCase,
	getOrdersUseCase *usecase.GetOrdersUseCase,
) *OrderHandler {
	return &OrderHandler{
		uploadOrderUseCase: uploadOrderUseCase,
		getOrdersUseCase:   getOrdersUseCase,
	}
}

func (h *OrderHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	resp, err := h.uploadOrderUseCase.Execute(r.Context(), usecase.UploadOrderRequest{
		UserID: userID,
		Number: orderNumber,
	})
	if err != nil {
		if err.Error() == "invalid order number format" {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if err.Error() == "order number already exists for another user" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if resp.Status == "already_uploaded" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.getOrdersUseCase.Execute(r.Context(), usecase.GetOrdersRequest{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}


