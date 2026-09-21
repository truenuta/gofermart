// Package model содержит структуры системы лояльности
package model

import "time"

// User — зарегистрированный пользователь системы лояльности
type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

type Order struct {
	Number     string
	Status     string
	Accrual    *float64
	UploadedAt time.Time
}
