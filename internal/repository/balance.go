package repository

import (
	"context"
	"errors"

	"github.com/truenuta/gofermart/internal/model"
)

var ErrNoAccrual = errors.New("no accrual")

func (r *UserRepository) SumAccrualByUser(ctx context.Context, userID int64) (float64, error) {
	sqlAccrual := "SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'"
	var accrual float64
	err := r.db.QueryRowContext(ctx, sqlAccrual, userID).Scan(&accrual)
	if err != nil {
		return 0, err
	}
	return accrual, nil

}

func (r *UserRepository) SumWithdrawalsByUser(ctx context.Context, userID int64) (float64, error) {
	sqlWithdrawals := "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1"
	var withdrawals float64
	err := r.db.QueryRowContext(ctx, sqlWithdrawals, userID).Scan(&withdrawals)
	if err != nil {
		return 0, err
	}
	return withdrawals, nil
}

func (r *UserRepository) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT id FROM users WHERE id = $1 FOR UPDATE", userID); err != nil {
		return err
	}
	var balance float64
	err = tx.QueryRowContext(ctx, `
          SELECT
            (SELECT COALESCE(SUM(accrual),0) FROM orders WHERE user_id=$1 AND status='PROCESSED')
          - (SELECT COALESCE(SUM(sum),0)     FROM withdrawals WHERE user_id=$1)`, userID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance < sum {
		return ErrNoAccrual
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO withdrawals (order_number, user_id, sum) VALUES ($1,$2,$3)", order, userID, sum); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepository) ListWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	var withdrawals []model.Withdrawal
	var withdrawal model.Withdrawal
	sql := "SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC"
	rows, err := r.db.QueryContext(ctx, sql, userID)
	if err != nil {
		return withdrawals, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(
			&withdrawal.Order,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)

	}
	if err := rows.Err(); err != nil {
		return withdrawals, err
	}
	return withdrawals, nil
}
