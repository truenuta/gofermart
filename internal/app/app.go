package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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

	r := chi.NewRouter()
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.With(handler.Authenticate(tokenManager)).Post("/api/user/orders", ordersHandler.OrderUpload)
	r.With(handler.Authenticate(tokenManager)).Get("/api/user/orders", ordersHandler.ListOrders)

	r.With(handler.Authenticate(tokenManager)).Get("/api/user/balance", balanceHandler.GetBalance)

	r.With(handler.Authenticate(tokenManager)).Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
	r.With(handler.Authenticate(tokenManager)).Get("/api/user/withdrawals", balanceHandler.ListWithdrawalsByUser)

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	worker := service.NewAccrualWorker(accrualClient, repository, time.Second)
	go worker.Run(ctx)

	LaSerr := http.ListenAndServe(cfg.RunAddress, r)
	if LaSerr != nil {
		return LaSerr
	}
	return nil
}
