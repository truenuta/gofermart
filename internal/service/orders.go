package service

import (
	"context"
	"errors"

	"github.com/truenuta/gofermart/internal/model"
)

var ErrInvalidOrderNumber = errors.New("invalid order number")

type OrdersRepo interface {
	CreateOrder(ctx context.Context, number string, userID int64) error
	ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
}

type OrderService struct {
	repo OrdersRepo
}

func NewOrderService(repo OrdersRepo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

// isValidLuhn проверяет номер заказа алгоритмом Луна
func isValidLuhn(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		c := number[i]
		if c < '0' || c > '9' {
			return false
		}

		digit := int(c - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}

func (os *OrderService) UploadOrder(ctx context.Context, userID int64, number string) error {
	if number == "" || !isValidLuhn(number) {
		return ErrInvalidOrderNumber
	}
	err := os.repo.CreateOrder(ctx, number, userID)
	if err != nil {
		return err
	}
	return nil

}

func (os *OrderService) ListOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	orders, err := os.repo.ListOrdersByUser(ctx, userID)
	if err != nil {
		return []model.Order{}, err
	}
	return orders, nil
}
