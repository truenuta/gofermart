package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/truenuta/gofermart/internal/app"
	"github.com/truenuta/gofermart/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Не удалось получить конфиг", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = app.Run(ctx, cfg)
	if err != nil {
		log.Fatal("Не удалось запустить сервис", err)
	}

}
