package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/truenuta/gofermart/internal/model"
)

var ErrOtherUsersOrder = errors.New("Order nuber has already been created by other user")
var ErrOrderNotFound = errors.New("Order was not found")
var ErrAlreadyExists = errors.New("Order already exists")

func (r *UserRepository) CreateOrder(ctx context.Context, number string, userID int64) error {
	var num string
	checkOrdersql := "SELECT number FROM orders WHERE orders.number = $1 AND orders.user_id = $2 "
	err := r.db.QueryRowContext(ctx, checkOrdersql, number, userID).Scan(&num)
	switch {
	case err == nil:
		return ErrAlreadyExists
	case errors.Is(err, sql.ErrNoRows):
	default:
		return err
	}

	getIDsql := "INSERT INTO orders (number, user_id) VALUES ($1, $2) ON CONFLICT (number) DO UPDATE SET user_id = orders.user_id RETURNING user_id"
	row := r.db.QueryRowContext(ctx, getIDsql, number, userID)
	var returnedID int64
	err = row.Scan(&returnedID)

	if err != nil {
		return err
	}
	if userID != returnedID {
		return ErrOtherUsersOrder
	}
	return nil
}

func (r *UserRepository) ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	var order model.Order
	var orders []model.Order
	getIDsql := "SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC"
	rows, err := r.db.QueryContext(ctx, getIDsql, userID)
	if err != nil {
		return orders, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)

	}
	if err := rows.Err(); err != nil {
		return orders, err
	}
	return orders, nil

}

func (r *UserRepository) ListPendingOrders(ctx context.Context) ([]string, error) {
	getOrdersSql := "SELECT number FROM orders WHERE status IN ('NEW', 'PROCESSING') ORDER BY uploaded_at"
	rows, err := r.db.QueryContext(ctx, getOrdersSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []string
	var number string
	for rows.Next() {
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		orders = append(orders, number)
	}
	if err := rows.Err(); err != nil {
		return orders, err
	}
	return orders, nil
}

func (r *UserRepository) UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error {
	sqlUpdate := "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"
	_, err := r.db.ExecContext(ctx, sqlUpdate, status, accrual, number)
	if err != nil {
		return err
	}
	return nil
}
