package handler

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey int

type TokenManager interface {
	GetUserID(token string) (int64, error)
}

const userIDKey ctxKey = 0

// Authenticate — middleware: пропускает запрос дальше только с валидным
// JWT в заголовке Authorization, помещая ID пользователя в контекст запроса.
func Authenticate(tm TokenManager) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := tm.GetUserID(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
