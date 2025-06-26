package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"GophKeeper.ru/internal/entities"
	"github.com/golang-jwt/jwt/v5"
)

func Sha256hash(value string) string {
	hash := sha256.New()
	hash.Write([]byte(value))
	return hex.EncodeToString(hash.Sum(nil))
}

func LoginFromToken(sToken, secretKey string) (int, error) {
	claims := &entities.Claims{}
	tkn, err := jwt.ParseWithClaims(sToken, claims, func(jwtKey *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return -1, err
	}

	if !tkn.Valid {
		return -1, errors.New("no valid token")
	}

	return claims.IDUser, nil
}
