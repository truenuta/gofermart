package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/truenuta/gofermart/internal/auth"
)

func buildTestJWT(t *testing.T, userID int64, exp time.Time, secret string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		UserID: userID,
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to build test token: %v", err)
	}
	return signed
}

func TestAuthenticate(t *testing.T) {
	testUserID := int64(23)
	validToken, err := auth.BuildJWTString(testUserID)
	if err != nil {
		t.Errorf("failed to build valid token: %v", err)
	}
	expiredToken := buildTestJWT(t, testUserID, time.Now().Add(-time.Hour), "supersecretkey")
	wrongSignatureToken := buildTestJWT(t, testUserID, time.Now().Add(time.Hour), "wrongsecret")
	tests := []struct {
		name           string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
		wantUserID     int64
	}{
		{
			name:           "no authorization header",
			authHeader:     "",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "missing bearer prefix",
			authHeader:     "sometoken",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "bearer with empty token",
			authHeader:     "Bearer ",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "wrong signature",
			authHeader:     "Bearer " + wrongSignatureToken,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + expiredToken,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantUserID:     testUserID,
		},
	}
	for _, test := range tests {
		var (
			nextCalled bool
			gotUserID  int64
		)
		t.Run(test.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			rec := httptest.NewRecorder()
			Authenticate(handler).ServeHTTP(rec, req)
			if rec.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, test.wantStatus)
			}
			if nextCalled != test.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, test.wantNextCalled)
			}
			if test.wantNextCalled && gotUserID != test.wantUserID {
				t.Errorf("userID in context = %d, want %d", gotUserID, test.wantUserID)
			}
		})
	}
}
