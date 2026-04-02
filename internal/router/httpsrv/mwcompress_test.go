package httpsrv

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/require"
)

func TestGZipMiddleware(t *testing.T) {
	t.Run("read_compressed_request", func(t *testing.T) {
		var testData = `{"f": 1}`
		var handler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				resultData, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.JSONEq(t, testData, string(resultData))
			},
		)

		var srv = httptest.NewServer(gzipMiddleware(handler))
		defer srv.Close()

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(testData))
		require.NoError(t, err)
		require.NoError(t, zb.Close())

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer func() {
			if errClose := resp.Body.Close(); errClose != nil {
				println(errClose.Error())
			}
		}()
	})

	t.Run("write_compressed_response", func(t *testing.T) {
		var testData = []byte("test")
		var handler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write(testData); err != nil {
					t.Fatal(err.Error())
				}
				w.WriteHeader(http.StatusCreated)
			},
		)

		var srv = httptest.NewServer(gzipMiddleware(handler))
		defer srv.Close()

		r := httptest.NewRequest("POST", srv.URL, nil)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Content-Type", "text/html")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		defer func() {
			if errClose := resp.Body.Close(); errClose != nil {
				println(errClose.Error())
			}
		}()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		resultData, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, testData, resultData)
	})
}

func TestGZipMiddlewareErrors(t *testing.T) {
	t.Run("reader_error", func(t *testing.T) {
		patchCBR := monkey.Patch(comressedBodyReader,
			func(io.ReadCloser) (*compressedDataReader, error) {
				return nil, fmt.Errorf("errTest")
			})
		defer patchCBR.Unpatch()

		var handler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				//nolint:funlen
			},
		)

		var srv = httptest.NewServer(gzipMiddleware(handler))
		defer srv.Close()

		r := httptest.NewRequest("POST", srv.URL, nil)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Type", "text/html")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		defer func() {
			if errClose := resp.Body.Close(); errClose != nil {
				println(errClose.Error())
			}
		}()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("wrong_compress_method", func(t *testing.T) {
		var handler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				//nolint:funlen
			},
		)

		var srv = httptest.NewServer(gzipMiddleware(handler))
		defer srv.Close()

		r := httptest.NewRequest("POST", srv.URL, nil)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "fake")
		r.Header.Set("Content-Type", "text/html")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		defer func() {
			if errClose := resp.Body.Close(); errClose != nil {
				println(err.Error())
			}
		}()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	})
}

func TestCompDataWriterHeader(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("test", "1")
	cw := compressedDataWriter{w: w}
	require.Equal(t, "1", cw.Header().Get("test"))
}
