// Package handler содержит HTTP-обработчики API системы лояльности.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
)

// credentials — тело запроса регистрации/входа.
type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Authenticator — методы сервиса аутентификации, нужные AuthHandler.
type Authenticator interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// AuthHandler обслуживает регистрацию и вход пользователей.
type AuthHandler struct {
	auth Authenticator
}

// NewAuthHandler создаёт AuthHandler поверх сервиса аутентификации.
func NewAuthHandler(auth Authenticator) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

// Register обрабатывает POST /api/user/register.
func (au *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	token, err := au.auth.Register(r.Context(), c.Login, c.Password)

	switch {
	case err == nil:
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidRequest):
		http.Error(w, "bad request", http.StatusBadRequest)
	case errors.Is(err, repository.ErrLoginTaken):
		http.Error(w, "login already taken", http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

}

// Login обрабатывает POST /api/user/login.
func (au *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	token, err := au.auth.Login(r.Context(), c.Login, c.Password)

	switch {
	case err == nil:
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidRequest):
		http.Error(w, "bad request", http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCredentials):
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

}
