package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/truenuta/gofermart/internal/accrual"
	"golang.org/x/sync/errgroup"
)

type AccrualClient interface {
	GetOrder(ctx context.Context, number string) (*accrual.Result, error)
}

type AccrualRepo interface {
	ListPendingOrders(ctx context.Context) ([]string, error)
	UpdateOrderAccrual(ctx context.Context, number, status string, accrual *float64) error
}

type AccrualWorker struct {
	client   AccrualClient
	repo     AccrualRepo
	interval time.Duration
	workers  int
}

func NewAccrualWorker(client AccrualClient, repo AccrualRepo, interval time.Duration, workers int) *AccrualWorker {
	return &AccrualWorker{
		client:   client,
		repo:     repo,
		interval: interval,
		workers:  workers,
	}
}

func (w *AccrualWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *AccrualWorker) processOrder(ctx context.Context, order string) error {
	res, err := w.client.GetOrder(ctx, order)
	var tooMany *accrual.TooManyRequestsError

	switch {
	case err == nil:
	case errors.Is(err, accrual.ErrOrderNotRegistered):
		return nil
	case errors.As(err, &tooMany):
		return tooMany
	default:
		log.Printf("accrual worker: get order %s: %v", order, err)
		return nil
	}
	ourStstus := mapStatus(res.Status)
	updateErr := w.repo.UpdateOrderAccrual(ctx, order, ourStstus, res.Accrual)
	if updateErr != nil {
		log.Printf("accrual worker: cant upload order: %v", updateErr)
	}
	return nil
}

func (w *AccrualWorker) processBatch(ctx context.Context) {
	orders, err := w.repo.ListPendingOrders(ctx)
	if err != nil {
		log.Printf("accrual worker: list pending orders: %v", err)
		return
	}
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(w.workers)
	for _, order := range orders {
		if gCtx.Err() != nil {
			break
		}
		g.Go(func() error {
			return w.processOrder(gCtx, order)
		})

	}
	err = g.Wait()
	var tooMany *accrual.TooManyRequestsError
	if errors.As(err, &tooMany) {
		sleepCtx(ctx, tooMany.RetryAfter)
	}

}

func mapStatus(status string) string {
	if status == "REGISTERED" {
		return "NEW"
	}
	return status
}

func sleepCtx(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}
