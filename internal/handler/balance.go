package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (float64, float64, error)
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
