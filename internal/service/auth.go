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

type TokenManager interface {
	BuildJWTString(userID int64) (string, error)
}

type AuthService struct {
	repo  UserRepo
	token TokenManager
}

func NewAuthService(repo UserRepo, token TokenManager) *AuthService {
	return &AuthService{
		repo:  repo,
		token: token,
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
	return as.token.BuildJWTString(id)

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
	return as.token.BuildJWTString(user.ID)
}
