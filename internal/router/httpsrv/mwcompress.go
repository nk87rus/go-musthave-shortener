package httpsrv

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// reader
		ce := getHeadderValues(r, "Content-Encoding")
		if slices.Contains(ce, "gzip") {
			bodyReader, err := comressedBodyReader(r.Body)
			if err != nil {
				log.Err(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = bodyReader
			defer func() {
				if err := bodyReader.Close(); err != nil {
					log.Err(err)
				}
			}()
		} else {
			if ce != nil {
				err := fmt.Errorf("метод сжатия %q не поддерживается", strings.Join(ce, ", "))
				log.Err(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		// writer
		var (
			writer      = w
			contentType = r.Header.Get("Content-Type")
		)
		if strings.EqualFold(contentType, "application/json") || strings.EqualFold(contentType, "text/html") {
			if slices.Contains(getHeadderValues(r, "Accept-Encoding"), "gzip") {
				compWriter := compressedRespWriter(w)
				writer = compWriter
				defer func() {
					if err := compWriter.Close(); err != nil {
						log.Err(err)
					}
				}()
			}
		}

		next.ServeHTTP(writer, r)
	})
}

func getHeadderValues(r *http.Request, headerName string) []string {
	var data = r.Header.Get(headerName)
	if data == "" {
		return nil
	}
	return strings.Split(data, ",")
}

// ---- compReader
type compressedDataReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func comressedBodyReader(r io.ReadCloser) (*compressedDataReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressedDataReader{r: r, zr: zr}, nil
}

func (c compressedDataReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressedDataReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// ---- copWriter

type compressedDataWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func compressedRespWriter(w http.ResponseWriter) *compressedDataWriter {
    gz := gzipPool.Get().(*gzip.Writer)
    gz.Reset(w)

	return &compressedDataWriter{w: w, zw: gz}
}


func (c *compressedDataWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressedDataWriter) Write(p []byte) (int, error) {
	wi, werr := c.zw.Write(p)
	return wi, werr
}

func (c *compressedDataWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressedDataWriter) Close() error {
	gzipPool.Put(c.zw)
	return c.zw.Close()
}
