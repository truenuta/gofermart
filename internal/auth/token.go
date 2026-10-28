package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// срок жизни выпускаемого токена
const TokenExp = time.Hour * 3

// secretKey - ключ подписи JWT
type TokenManager struct {
	secretKey []byte
	ttl       time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secretKey: []byte(secret),
		ttl:       ttl,
	}
}

// Claims - полезная нагрузка JWT: стандартные поля плюс идентификатор пользователя
type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

// BuildJWTString выпускает подписанный JWT для пользователя UserID
func (tm *TokenManager) BuildJWTString(UserID int64) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.ttl)),
			},
			UserID: UserID,
		})
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

// GetUserID проверяет подпись и срок действия токена и возвращает ID пользователя
func (tm *TokenManager) GetUserID(tokenString string) (int64, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tm.secretKey), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}

	return claims.UserID, nil
}
