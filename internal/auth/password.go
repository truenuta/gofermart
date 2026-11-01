// Package auth отвечает за хеширование паролей и выпуск/проверку JWT.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword возвращает bcrypt-хеш пароля.
func HashPassword(pass string) (string, error) {
	hashString, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hashString), err
}

// CheckPassword сверяет пароль с хешем; nil означает совпадение.
func CheckPassword(hash, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}
