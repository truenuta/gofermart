package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

// Config хранит параметры запуска сервиса. Каждый параметр может быть задан
// флагом командной строки либо одноимённой переменной окружения
type Config struct {
	// RunAddress - адрес и порт, на которых слушает HTTP-сервер
	// (флаг -a, переменная окружения RUN_ADDRESS)
	RunAddress string `env:"RUN_ADDRESS"`
	// DatabaseURI — строка подключения к PostgreSQL
	// (флаг -d, переменная окружения DATABASE_URI)
	DatabaseURI string `env:"DATABASE_URI"`
	// AccrualSystemAddress — базовый адрес внешней системы расчёта начислений
	// (флаг -r, переменная окружения ACCRUAL_SYSTEM_ADDRESS).
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

// ParseFlags разбирает флаги командной строки
func ParseFlags() *Config {
	var RunAddress string
	var DatabaseURI string
	var AccuralSystemAddress string

	flag.StringVar(&RunAddress, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&DatabaseURI, "d", "", "адрес подключения к базе данных")
	flag.StringVar(&AccuralSystemAddress, "r", "", "адрес системы расчёта начислений")
	flag.Parse()

	return &Config{
		RunAddress:           RunAddress,
		DatabaseURI:          DatabaseURI,
		AccrualSystemAddress: AccuralSystemAddress,
	}

}

// ParseEnv читает переменные окружения
func ParseEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewConfig собирает итоговую конфигурацию сервиса
func NewConfig() (*Config, error) {
	flags := ParseFlags()
	envCnfg, err := ParseEnv()
	if err != nil {
		return nil, err
	}
	if envCnfg.RunAddress == "" {
		envCnfg.RunAddress = flags.RunAddress
	}
	if envCnfg.DatabaseURI == "" {
		envCnfg.DatabaseURI = flags.DatabaseURI
	}
	if envCnfg.AccrualSystemAddress == "" {
		envCnfg.AccrualSystemAddress = flags.AccrualSystemAddress
	}
	return envCnfg, nil

}
