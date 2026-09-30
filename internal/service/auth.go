package service

import (
	"context"
	"errors"

	"github.com/truenuta/gofermart/internal/auth"
	"github.com/truenuta/gofermart/internal/model"
	"github.com/truenuta/gofermart/internal/repository"
)

var ErrInvalidRequest = errors.New("invalid request")
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepo interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
}

type AuthService struct {
	repo UserRepo
}

func NewAuthService(repo UserRepo) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (as *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrInvalidRequest
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	id, err := as.repo.CreateUser(ctx, login, hash)
	if err != nil {
		return "", err
	}
	return auth.BuildJWTString(id)

}

func (as *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrInvalidRequest
	}
	user, err := as.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	return auth.BuildJWTString(user.ID)
}
