package httpsrv

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	respData struct {
		status, size int
	}


	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *respData
	}
)

func (r *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var (
			rd        = respData{}
			logWriter = loggerResponseWriter{
				responseData:   &rd,
				ResponseWriter: w,
			}
		)

		next.ServeHTTP(&logWriter, r)

		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Dur("duration", time.Since(start)).
			Int("status", rd.status).
			Int("size", rd.size).
			Msg("Новый запрос")
	})
}
