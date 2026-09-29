package app

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/truenuta/gofermart/internal/config"
	"github.com/truenuta/gofermart/internal/db"
	"github.com/truenuta/gofermart/internal/handler"
	"github.com/truenuta/gofermart/internal/repository"
	"github.com/truenuta/gofermart/internal/service"
)

// Run регистрирует HTTP-маршруты API и запускает сервер
func Run(cfg *config.Config) error {
	database, err := db.NewDB(cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}
	defer database.Close()
	err = db.RunMigrations(database)
	if err != nil {
		return fmt.Errorf("миграции отдали ошибку: %w", err)
	}
	repository := repository.NewUserRepository(database)

	authService := service.NewAuthService(repository)
	authHandler := handler.NewAuthHandler(authService)

	ordersService := service.NewOrderService(repository)
	ordersHandler := handler.NewOrdersHandler(ordersService)

	balanceService := service.NewBalanceService(repository)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	r := chi.NewRouter()
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.With(handler.Authenticate).Post("/api/user/orders", ordersHandler.OrderUpload)
	r.With(handler.Authenticate).Get("/api/user/orders", ordersHandler.ListOrders)

	r.With(handler.Authenticate).Get("/api/user/balance", balanceHandler.GetBalance)

	r.With(handler.Authenticate).Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
	r.With(handler.Authenticate).Get("/api/user/withdrawals", balanceHandler.ListWithdrawalsByUser)

	LaSerr := http.ListenAndServe(cfg.RunAddress, r)
	if LaSerr != nil {
		return LaSerr
	}
	return nil
}
