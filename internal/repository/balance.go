package repository

import "context"

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
