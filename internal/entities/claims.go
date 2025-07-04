package entities

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	IDUser int `json:"id_user"`
	jwt.RegisteredClaims
}

type User struct {
	ID        int    `json:"id,omitempty"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	IsDisable bool   `json:"is_disable,omitempty"`
}

// IsEcual сравнение двух пользователей
func (u *User) IsEcual(user *User) bool {
	return u.Login == user.Login && u.Password == user.Password
}

// GetToken создает токен для пользователя
func GetToken(u *User) (string, error) {
	expirationTime := time.Now().Add(time.Hour)
	claims := &Claims{
		IDUser: u.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		slog.Error(fmt.Sprintf("getToken %s", err))
		return "", err
	}

	return tokenString, nil
}
