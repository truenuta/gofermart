package service

import "context"

type BalanceRepo interface {
	SumAccrualByUser(ctx context.Context, userID int64) (float64, error)
	SumWithdrawalsByUser(ctx context.Context, userID int64) (float64, error)
}

type BalanceService struct {
	repo BalanceRepo
}

func NewBalanceService(repo BalanceRepo) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

func (bs *BalanceService) GetBalance(ctx context.Context, userID int64) (float64, float64, error) {
	accrual, err := bs.repo.SumAccrualByUser(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	withdrawals, err := bs.repo.SumWithdrawalsByUser(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	sum := accrual - withdrawals
	return sum, withdrawals, nil
}
