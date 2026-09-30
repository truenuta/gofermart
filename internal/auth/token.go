package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// срок жизни выпускаемого токена
const TokenExp = time.Hour * 3

// secretKey - ключ подписи JWT
const secretKey = "supersecretkey" // just for autotest and study purpose, not real key

// Claims - полезная нагрузка JWT: стандартные поля плюс идентификатор пользователя
type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

// BuildJWTString выпускает подписанный JWT для пользователя UserID
func BuildJWTString(UserID int64) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
			UserID: UserID,
		})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

// GetUserID проверяет подпись и срок действия токена и возвращает ID пользователя
func GetUserID(tokenString string) (int64, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}

	return claims.UserID, nil
}
