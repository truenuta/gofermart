package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/truenuta/gofermart/internal/auth"
)

type ctxKey int

const userIDKey ctxKey = 0

// Authenticate — middleware: пропускает запрос дальше только с валидным
// JWT в заголовке Authorization, помещая ID пользователя в контекст запроса.
func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, err := auth.GetUserID(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
