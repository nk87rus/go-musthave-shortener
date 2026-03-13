package httpsrv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/filestorage"
	memstorage "github.com/nk87rus/go-musthave-shortener/internal/repository/mem"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	var errBaseURL = fmt.Errorf("не корректный base address")

	testCases := []struct {
		name       string
		addr       string
		baddr      string
		wantResult Server
		wantError  error
	}{
		{
			name:      "errBaseURL",
			baddr:     ":abc.xx",
			wantError: errBaseURL,
		},
		{
			name:  "Correct",
			addr:  "127.0.0.1:8080",
			baddr: "http://127.0.0.2:8090",
			wantResult: Server{
				addr:     "127.0.0.1:8080",
				baseURL:  &url.URL{Scheme: "http", Host: "127.0.0.2:8090"},
				handlers: new(handler.Handlers),
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultData, resultError := New(tc.addr, tc.baddr, nil, nil)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &Server{}, resultData)
			}
		})
	}
}

func TestServerRun(t *testing.T) {
	s := &Server{addr: "test", baseURL: &url.URL{}}
	go func() {
		if err := s.Run(t.Context()); err != nil {
			println(err.Error())
		}
	}()
}

func TestDelURLs(t *testing.T) {
	testCases := []struct {
		name      string
		body      []byte
		wantError error
	}{
		{
			name:      "errDecode",
			wantError: io.EOF,
		},
		{
			name:      "noURL",
			body:      []byte(`[]`),
			wantError: fmt.Errorf("список URL для удаления пуст"),
		},
		{
			name:      "Correct",
			body:      []byte(`["a", "b"]`),
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hMock := NewMockHandlers(t)
			hMock.On("DelURLs", mock.Anything, mock.Anything).Return().Maybe()

			s := &Server{handlers: hMock}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/", bytes.NewBuffer(tc.body))
			s.delURLs(w, r)
			if tc.wantError != nil {
				require.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				require.Equal(t, http.StatusAccepted, w.Code)
			}
		})
	}
}

func ExampleServer_Run() {
	// Server

	tmpFile, err := os.CreateTemp("", "sample-*.json")
	if err != nil {
		fmt.Println("error on create temporary file:", err.Error())
		return
	}
	defer func() {
		if errRemove := os.Remove(tmpFile.Name()); err != nil {
			println(errRemove.Error())
		}
	}()

	tmpExtStorage, err := filestorage.NewStorage(tmpFile.Name())
	if err != nil {
		fmt.Println("error on init file storage:", err.Error())
		return
	}

	ctx := context.Background()
	tmpStorage, err := memstorage.NewStorage(ctx, tmpExtStorage)
	if err != nil {
		fmt.Println("error on init storage:", err.Error())
		return
	}

	addr := url.URL{Scheme: "http", Host: "localhost:8100"}
	s, err := New(addr.Host, addr.String(), tmpStorage, nil)
	if err != nil {
		println("error on init http server:", err.Error())
	}

	go func() {
		if errRun := s.Run(ctx); err != nil {
			println(errRun.Error())
		}
	}()

	time.Sleep(time.Second)

	// Client
	client := resty.New()
	baseURL := "http://test.ru"

	// Create short URL
	req := addr.JoinPath("/").String()
	resp, err := client.R().SetContext(ctx).SetBody(baseURL).Post(req)
	if err != nil {
		fmt.Println("error on create short URL:", err.Error())
	}
	shortURL, _ := url.Parse(resp.String())
	fmt.Println(resp.StatusCode())

	// Restore baseURL
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	req = addr.JoinPath(shortURL.Path).String()
	resp, err = client.R().SetContext(ctx).Get(req)
	if err != nil && !strings.Contains(err.Error(), "auto redirect is disabled") {
		fmt.Println("error on restore base URL:", err.Error())
	}
	fmt.Println(resp.StatusCode())

	// Output:
	// 201
	// 307
}
