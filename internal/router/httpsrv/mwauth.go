package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/rs/zerolog/log"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	TokenExp   = time.Hour
	CookieName = "uid"
	AuthHeader = "Authorization"
)

var (
	key             string
	ErrTokenInvalid = fmt.Errorf("токен не валиден")
)

func init() {
	if envKey, ok := os.LookupEnv("KEY"); ok {
		key = envKey
	} else {
		log.Warn().Msg("ключ не указан, используется тестовый ключ")
		key = "super_secret_key"
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := getUIDCookie(r)
		if err != nil {
			log.Err(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ahValue := r.Header.Get(AuthHeader)
		fmt.Printf("DEBUG authMiddleware received:\n\tcookie: %+v\n\theader: %v\n", cookie, ahValue)

		if cookie == nil || !validateCookie(cookie) {
			// newCookie, token, err := makeCookie(ahValue)
			newCookie, _, err := makeCookie(ahValue)
			if err != nil {
				log.Err(err).Msg("ошибка при создании cookie")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			http.SetCookie(w, newCookie)

			// if ahValue == "" {
			// 	w.Header().Set(AuthHeader, token)
			// }
		}

		var userID string
		if cookie != nil {
			if uid, err := getCookieUserID(cookie); err != nil {
				log.Err(err)
				fmt.Printf("DEBUG authMiddleware: getCookieUserID: ERROR: %+v\n", err)
				// http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				userID = uid
			}
			// } else {
			// 	if uid, err := parseJWT(ahValue); err != nil {
			// 		log.Err(err)
			// 		fmt.Printf("DEBUG authMiddleware: header: ERROR: %+v\n", err)
			// 		// http.Error(w, err.Error(), http.StatusBadRequest)
			// 	} else {
			// 		userID = uid
			// 	}
		}

		if !validateCookie(cookie) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		ctx := context.WithValue(r.Context(), model.CtxUserID, userID)
		fmt.Printf("DEBUG authMiddleware: userID: %+v\n", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUIDCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		fmt.Printf("DEBUG getUIDCookie: err: %+v\n", err)
		log.Err(err).Msg("ошибка при получении cookie")
		return nil, err
	}
	return cookie, nil
}

func validateCookie(cookie *http.Cookie) bool {
	_, err := getCookieUserID(cookie)
	if err != nil {
		fmt.Printf("DEBUG validateCookie: err: %+v\n", err)
		return !errors.Is(err, ErrTokenInvalid)
	}
	return true
}

func getCookieUserID(cookie *http.Cookie) (string, error) {
	if cookie == nil {
		return "", fmt.Errorf("пустой cookie (cookie is nil)")
	}
	return parseJWT(cookie.Value)
}

func parseJWT(jwtToken string) (string, error) {
	if jwtToken == "" {
		return "", fmt.Errorf("пустой токен авторизации")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims,
		func(t *jwt.Token) (any, error) {
			return []byte(key), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrTokenInvalid
	}

	return claims.UserID, nil
}

func makeCookie(tokenValue string) (*http.Cookie, string, error) {
	if tokenValue == "" {
		newToken, err := makeJWT()
		if err != nil {
			return nil, "", err
		}
		tokenValue = newToken
	}

	return &http.Cookie{
			Name:     CookieName,
			Value:    tokenValue,
			Path:     "/",
			Expires:  time.Now().Add(TokenExp),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
		tokenValue,
		nil
}

func makeJWT() (string, error) {
	newUID := uuid.NewString()
	fmt.Printf("DEBUG makeJWT: newUID: %s\n", newUID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: newUID,
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
