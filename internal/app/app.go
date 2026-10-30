package app

import (
	"context"
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
)

// Run регистрирует HTTP-маршруты API и запускает сервер
func Run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
	worker := service.NewAccrualWorker(accrualClient, repository, time.Second)
	go worker.Run(ctx)

	LaSerr := http.ListenAndServe(cfg.RunAddress, mux)
	if LaSerr != nil {
		return LaSerr
	}
	return nil
}
