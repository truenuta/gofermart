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
	r := chi.NewRouter()
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)
	r.Post("/api/user/orders", notImplemented)
	r.Post("/api/user/balance/withdraw", notImplemented)
	r.Get("/api/user/orders", notImplemented)
	r.Get("/api/user/balance", notImplemented)
	r.Get("/api/user/withdrawals", notImplemented)
	LaSerr := http.ListenAndServe(cfg.RunAddress, r)
	if LaSerr != nil {
		return LaSerr
	}
	return nil
}

// notImplemented — временная заглушка для ещё не реализованных маршрутов.
// Отвечает статусом 501 Not Implemented.
func notImplemented(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
