package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/truenuta/gofermart/internal/accrual"
	"github.com/truenuta/gofermart/internal/auth"
	"github.com/truenuta/gofermart/internal/config"
	"github.com/truenuta/gofermart/internal/db"
	"github.com/truenuta/gofermart/internal/handler"
	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
	"golang.org/x/sync/errgroup"
)

// accrualWorkers — сколько заказов воркер одновременно запрашивает в системе начислений
const accrualWorkers = 5

// Run регистрирует HTTP-маршруты API и запускает сервер
func Run(ctx context.Context, cfg *config.Config) error {
	database, err := db.NewDB(ctx, cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}
	defer database.Close()
	err = db.RunMigrations(database)
	if err != nil {
		return fmt.Errorf("миграции отдали ошибку: %w", err)
	}
	repository := repository.NewUserRepository(database)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, auth.TokenExp)

	authService := service.NewAuthService(repository, tokenManager)
	authHandler := handler.NewAuthHandler(authService)

	ordersService := service.NewOrderService(repository)
	ordersHandler := handler.NewOrdersHandler(ordersService)

	balanceService := service.NewBalanceService(repository)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/user/register", authHandler.Register)
	mux.HandleFunc("POST /api/user/login", authHandler.Login)

	authMW := handler.Authenticate(tokenManager)

	mux.Handle("POST /api/user/orders", authMW(http.HandlerFunc(ordersHandler.OrderUpload)))
	mux.Handle("GET /api/user/orders", authMW(http.HandlerFunc(ordersHandler.ListOrders)))

	mux.Handle("GET /api/user/balance", authMW(http.HandlerFunc(balanceHandler.GetBalance)))

	mux.Handle("POST /api/user/balance/withdraw", authMW(http.HandlerFunc(balanceHandler.Withdraw)))
	mux.Handle("GET /api/user/withdrawals", authMW(http.HandlerFunc(balanceHandler.ListWithdrawalsByUser)))

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	worker := service.NewAccrualWorker(accrualClient, repository, time.Second, accrualWorkers)

	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		worker.Run(gCtx)
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
