package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/truenuta/gofermart/internal/model"
	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
)

type orderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type OrdersService interface {
	UploadOrder(ctx context.Context, userID int64, number string) error
	ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}

type OrdersHandler struct {
	orders OrdersService
}

func NewOrdersHandler(os OrdersService) *OrdersHandler {
	return &OrdersHandler{
		orders: os,
	}
}

func (oh *OrdersHandler) OrderUpload(w http.ResponseWriter, r *http.Request) {
	numberBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	number := string(numberBody)
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = oh.orders.UploadOrder(r.Context(), userID, number)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, repository.ErrAlreadyExists):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, repository.ErrOtherUsersOrder):
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, service.ErrInvalidOrderNumber):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (oh *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	orders, err := oh.orders.ListOrders(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		response = append(response, orderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		})
	}
	json.NewEncoder(w).Encode(response)
}
