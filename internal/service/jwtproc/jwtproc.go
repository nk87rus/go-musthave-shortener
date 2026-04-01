package jwtproc

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type JWTProc struct {
	key string
}

const tokenExp = time.Hour

var ErrTokenInvalid = fmt.Errorf("токен не валиден")

func New() *JWTProc {
	var key string
	if envKey, ok := os.LookupEnv("KEY"); ok {
		key = envKey
	} else {
		log.Warn().Str("server", "grpc").Msg("ключ не указан, используется тестовый ключ")
		key = "super_secret_key"
	}

	return &JWTProc{key: key}
}

func (jp *JWTProc) TokenExpiration() time.Duration {
	return tokenExp
}

func (jp *JWTProc) ParseJWT(jwtToken string) (string, error) {
	if jwtToken == "" {
		return "", fmt.Errorf("пустой токен авторизации")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims,
		func(t *jwt.Token) (any, error) {
			return []byte(jp.key), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrTokenInvalid
	}

	return claims.UserID, nil
}

func (jp *JWTProc) MakeJWT() (string, error) {
	newUID := uuid.NewString()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: newUID,
	})

	tokenString, err := token.SignedString([]byte(jp.key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
