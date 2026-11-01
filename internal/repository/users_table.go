// Package repository даёт доступ к данным системы лояльности в PostgreSQL.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/truenuta/gofermart/internal/model"
)

// ErrLoginTaken возвращается, когда логин уже занят другим пользователем
var ErrLoginTaken = errors.New("login already taken")

// ErrUserNotFound возвращается, когда пользователь с указанным логином не найден
var ErrUserNotFound = errors.New("user not found")

// UserRepository хранит и читает пользователей в таблице users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт репозиторий поверх открытого пула соединений
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser сохраняет нового пользователя и возвращает его ID
// Если логин уже занят, возвращает ErrLoginTaken
func (r *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	getIDsql := "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id"
	row := r.db.QueryRowContext(ctx, getIDsql, login, passwordHash)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, ErrLoginTaken
		}
		return 0, err
	}
	return id, nil
}

// GetUserByLogin возвращает пользователя по логину
// Если пользователя нет, возвращает ErrUserNotFound
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	sqlFindLogin := "SELECT id, login, password_hash FROM users WHERE login = $1"
	row := r.db.QueryRowContext(ctx, sqlFindLogin, login)
	var user model.User
	err := row.Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}
	return user, nil
}
