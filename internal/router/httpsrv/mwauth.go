package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/nk87rus/go-musthave-shortener/internal/service/jwtproc"
	"github.com/rs/zerolog/log"
)

const (
	CookieName = "uid"
	AuthHeader = "Authorization"
)

type JWTProcessor interface {
	ParseJWT(jwtToken string) (string, error)
	MakeJWT() (string, error)
	TokenExpiration() time.Duration
}

var jwtProc JWTProcessor

func init() {
	jwtProc = jwtproc.New()
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

		var userID string
		if cookie == nil || !validateCookie(cookie) {
			newCookie, tv, err := makeCookie(ahValue)
			if err != nil {
				log.Err(err).Msg("ошибка при создании cookie")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			http.SetCookie(w, newCookie)
			w.Header().Set(AuthHeader, tv)
		}

		if ahValue != "" {
			if uid, err := jwtProc.ParseJWT(ahValue); err != nil {
				log.Err(err)
			} else {
				userID = uid
			}
		} else {
			if r.RequestURI != "/api/user/urls" {
				userID = model.AnonymousUserID
			}
		}

		ctx := context.WithValue(r.Context(), model.CtxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUIDCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		log.Err(err).Msg("ошибка при получении cookie")
		return nil, err
	}
	return cookie, nil
}

func validateCookie(cookie *http.Cookie) bool {
	_, err := getCookieUserID(cookie)
	if err != nil {
		return !errors.Is(err, jwtproc.ErrTokenInvalid)
		// return false
	}
	return true
}

func getCookieUserID(cookie *http.Cookie) (string, error) {
	if cookie == nil {
		return "", fmt.Errorf("пустой cookie (cookie is nil)")
	}
	return jwtProc.ParseJWT(cookie.Value)
}

func makeCookie(value string) (*http.Cookie, string, error) {
	var newTokenValue = value
	if newTokenValue == "" {
		newToken, err := jwtProc.MakeJWT()
		if err != nil {
			return nil, "", err
		}
		newTokenValue = newToken
	}

	return &http.Cookie{
			Name:     CookieName,
			Value:    newTokenValue,
			Path:     "/",
			Expires:  time.Now().Add(jwtProc.TokenExpiration()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
		newTokenValue,
		nil
}
