package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testOrder = "123456789222"

func newFakeAccural(t *testing.T, status int, body string, headers map[string]string) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/orders/"+testOrder {
			t.Errorf("неправильно передан запрос")
		}
		for key, value := range headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(status)
		if body != "" {
			w.Write([]byte(body))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetOrder_Processed(t *testing.T) {
	headers := map[string]string{"Content-Type": "application/json"}
	body := `{"order":"` + testOrder + `","status":"PROCESSED","accrual":500}`
	srv := newFakeAccural(t, http.StatusOK, body, headers)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if res.Order != testOrder {
		t.Errorf("order = %q, want %q", res.Order, testOrder)
	}
	if res.Status != "PROCESSED" {
		t.Errorf("status = %q, want %q", res.Status, "PROCESSED")
	}
	if res.Accrual == nil {
		t.Fatalf("accrual = nil, want 500")
	}

}

func TestGetOrder_Processing(t *testing.T) {
	headers := map[string]string{"Content-Type": "application/json"}
	body := `{"order":"` + testOrder + `","status":"PROCESSING"}`
	srv := newFakeAccural(t, http.StatusOK, body, headers)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if res.Order != testOrder {
		t.Errorf("order = %q, want %q", res.Order, testOrder)
	}
	if res.Status != "PROCESSING" {
		t.Errorf("status = %q, want %q", res.Status, "PROCESSING")
	}
	if res.Accrual != nil {
		t.Fatalf("accrual != nil, want nil")
	}

}

func TestGetOrder_NotRegistered(t *testing.T) {
	headers := map[string]string{"Content-Type": "application/json"}
	body := ``
	srv := newFakeAccural(t, http.StatusNoContent, body, headers)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if !errors.Is(err, ErrOrderNotRegistered) {
		t.Fatalf("error != ErrOrderNotRegistered, %v", err)

	}
	if res != nil {
		t.Fatalf("res should be empty!")
	}

}

func TestGetOrder_TooManyRequests(t *testing.T) {
	headers := map[string]string{"Retry-After": "30"}
	body := ``
	var tmErr *TooManyRequestsError
	srv := newFakeAccural(t, http.StatusTooManyRequests, body, headers)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if !errors.As(err, &tmErr) {
		t.Fatalf("error != TooManyRequestsError, %v", err)

	}
	if tmErr.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter = %v, want %v", tmErr.RetryAfter, 30*time.Second)
	}
	if res != nil {
		t.Errorf("res = %+v, want nil", res)
	}

}

func TestGetOrder_TooManyRequestsNoHeader(t *testing.T) {
	srv := newFakeAccural(t, http.StatusTooManyRequests, "", nil)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	var tooMany *TooManyRequestsError
	if !errors.As(err, &tooMany) {
		t.Fatalf("error = %v, want TooManyRequestsError", err)
	}
	if tooMany.RetryAfter != defaultRetryAfter {
		t.Errorf("RetryAfter = %v, want %v", tooMany.RetryAfter, defaultRetryAfter)
	}
	if res != nil {
		t.Errorf("res = %+v, want nil", res)
	}
}

func TestGetOrder_ServerError(t *testing.T) {
	srv := newFakeAccural(t, http.StatusInternalServerError, "", nil)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if err == nil {
		t.Fatalf("error = nil, want error")
	}
	var tooMany *TooManyRequestsError
	if errors.Is(err, ErrOrderNotRegistered) || errors.As(err, &tooMany) {
		t.Errorf("error = %v, want generic unexpected status error", err)
	}
	if res != nil {
		t.Errorf("res = %+v, want nil", res)
	}
}

func TestGetOrder_BadJSON(t *testing.T) {
	headers := map[string]string{"Content-Type": "application/json"}
	srv := newFakeAccural(t, http.StatusOK, `{not json`, headers)

	res, err := NewClient(srv.URL).GetOrder(context.Background(), testOrder)
	if err == nil {
		t.Fatalf("error = nil, want decode error")
	}
	if res != nil {
		t.Errorf("res = %+v, want nil", res)
	}
}
