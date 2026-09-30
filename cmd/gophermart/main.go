package main

import (
	"log"

	"github.com/truenuta/gofermart/internal/app"
	"github.com/truenuta/gofermart/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Не удалось получить конфиг", err)
	}
	err = app.Run(cfg)
	if err != nil {
		log.Fatal("Не удалось запустить сервис", err)
	}
}
