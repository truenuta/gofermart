package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrOrderNotRegistered = errors.New("order not registered in accrual")

const defaultRetryAfter = 60 * time.Second

type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("accrual rate limit, retry after %s", e.RetryAfter)
}

// Result — ответ accrual по одному заказу.
type Result struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetOrder(ctx context.Context, number string) (*Result, error) {
	URL := c.baseURL + "/api/orders/" + number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	res := Result{}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
			return nil, fmt.Errorf("decode accrual response: %w", err)
		}
		return &res, nil
	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		retry := parseRetryAfter(response.Header.Get("Retry-After"))
		return nil, &TooManyRequestsError{RetryAfter: retry}
	default:
		return nil, fmt.Errorf("unexpected accrual status: %d", response.StatusCode)
	}
}
func parseRetryAfter(retryAfter string) time.Duration {
	sec, err := strconv.Atoi(retryAfter)
	if err != nil || sec <= 0 {
		return defaultRetryAfter
	}
	return time.Duration(sec) * time.Second
}
