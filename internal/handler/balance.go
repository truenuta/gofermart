package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/truenuta/gofermart/internal/model"
	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (float64, float64, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	ListWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type BalanceHandler struct {
	service BalanceService
}

func NewBalanceHandler(service BalanceService) *BalanceHandler {
	return &BalanceHandler{
		service: service,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userId, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sum, withdrawals, err := h.service.GetBalance(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   sum,
		Withdrawn: withdrawals,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidOrderNumber):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Is(err, repository.ErrNoAccrual):
		w.WriteHeader(http.StatusPaymentRequired)
	case errors.Is(err, service.ErrNegativeOrZeroSum):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (h *BalanceHandler) ListWithdrawalsByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	withdrawals, err := h.service.ListWithdrawalsByUser(r.Context(), userID)
	switch {
	case err == nil:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrNoWithdraws):
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type withdrawalResponse struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	resp := make([]withdrawalResponse, 0, len(withdrawals))
	for _, wd := range withdrawals {
		resp = append(resp, withdrawalResponse{
			Order:       wd.Order,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		})
	}
	json.NewEncoder(w).Encode(resp)

}
