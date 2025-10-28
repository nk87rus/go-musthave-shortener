package httpsrv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLRWWrite(t *testing.T) {
	w := httptest.NewRecorder()
	rd := respData{}
	lrw := loggerResponseWriter{
		ResponseWriter: w,
		responseData:   &rd,
	}

	resultSize, resultError := lrw.Write([]byte("test"))
	require.Nil(t, resultError)
	require.Equal(t, 4, resultSize)
	require.Equal(t, 4, rd.size)
}

func TestLRWWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rd := respData{}
	lrw := loggerResponseWriter{
		ResponseWriter: w,
		responseData:   &rd,
	}

	lrw.WriteHeader(100)
	require.Equal(t, 100, rd.status)
	require.Equal(t, 100, w.Code)
}

func TesstLoggerMiddleware(t *testing.T) {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//nolint:funlen
		},
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	loggerMiddleware(handler).ServeHTTP(w, r)
}
