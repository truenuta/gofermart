package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
)

// fakeAuth - фейковая реализация Authenticator для тестов хендлера.
type fakeAuth struct {
	registerToken string
	registerErr   error
	loginToken    string
	loginErr      error
}

func (f fakeAuth) Register(_ context.Context, _, _ string) (string, error) {
	return f.registerToken, f.registerErr
}

func (f fakeAuth) Login(_ context.Context, _, _ string) (string, error) {
	return f.loginToken, f.loginErr
}

func TestRegister_OK(t *testing.T) {
	h := NewAuthHandler(fakeAuth{registerToken: "token123"})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Authorization"); got != "Bearer token123" {
		t.Errorf("Authorization header = %q, want %q", got, "Bearer token123")
	}
}
func TestRegister_BadJSON(t *testing.T) {
	h := NewAuthHandler(fakeAuth{registerToken: "token123"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"not json"}`))
	w := httptest.NewRecorder()
	h.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegister_InvalidRequest(t *testing.T) {
	h := NewAuthHandler(fakeAuth{registerErr: service.ErrInvalidRequest})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusBadRequest)
	}

}

func TestRegister_LoginTaken(t *testing.T) {
	h := NewAuthHandler(fakeAuth{registerErr: repository.ErrLoginTaken})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Register(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestRegister_InternalError(t *testing.T) {
	h := NewAuthHandler(fakeAuth{registerErr: errors.New("internal error")})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Register(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestLogin_OK(t *testing.T) {
	h := NewAuthHandler(fakeAuth{loginToken: "token123"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusOK)
	}

}

func TestLogin_BadJSON(t *testing.T) {
	h := NewAuthHandler(fakeAuth{loginToken: "token123"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"not json"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := NewAuthHandler(fakeAuth{loginErr: service.ErrInvalidCredentials})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

}
func TestLogin_InternalError(t *testing.T) {
	h := NewAuthHandler(fakeAuth{loginErr: errors.New("internal error")})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"anya","password":"secret"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
