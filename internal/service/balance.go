package service

import (
	"context"
	"errors"

	"github.com/truenuta/gofermart/internal/model"
)

var ErrNegativeOrZeroSum = errors.New("cant withdrow negative or zero sum")
var ErrNoWithdraws = errors.New("cant find any withdrows")

type BalanceRepo interface {
	SumAccrualByUser(ctx context.Context, userID int64) (float64, error)
	SumWithdrawalsByUser(ctx context.Context, userID int64) (float64, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	ListWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error)
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

func (bs *BalanceService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	if sum <= 0 {
		return ErrNegativeOrZeroSum
	}
	ok := isValidLuhn(order)
	if !ok {
		return ErrInvalidOrderNumber
	}
	err := bs.repo.Withdraw(ctx, userID, order, sum)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BalanceService) ListWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	response, err := bs.repo.ListWithdrawalsByUser(ctx, userID)
	if err != nil {
		return []model.Withdrawal{}, err
	}
	if len(response) == 0 {
		return response, ErrNoWithdraws
	}
	return response, nil
}
