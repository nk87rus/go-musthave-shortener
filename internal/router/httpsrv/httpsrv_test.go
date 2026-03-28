//go:generate go run github.com/vektra/mockery/v2 --name=ReadCloser --output=./ --outpkg=httpsrv --filename=mock_ReadCloser_test.go --dir=$GOROOT/src/io
//go:generate go run github.com/vektra/mockery/v2 --name=Database --output=./ --outpkg=httpsrv --filename=mock_HdlrDatabase_test.go --dir=../../handler
package httpsrv

import (
	"bytes"
	"context"
	"errors"
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
	"github.com/nk87rus/go-musthave-shortener/internal/model"
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
			resultData, resultError := New(tc.addr, tc.baddr, "", false, nil, nil)
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

func TestEnableAudit(t *testing.T) {
	s := Server{}
	require.Nil(t, s.audit)

	s.EnableAudit("", "")
	require.NotNil(t, s.audit)
}

func TestServerRun(t *testing.T) {
	s := &Server{addr: "test", baseURL: &url.URL{}}
	go func() {
		if err := s.Run(t.Context()); err != nil {
			println(err.Error())
		}
	}()
}

func TestCreateShortURL(t *testing.T) {
	testCases := []struct {
		name       string
		req        *http.Request
		mFunc      func(m *MockHandlers)
		wantStatus int
	}{
		{
			name:       "WrongMethod",
			req:        httptest.NewRequest(http.MethodDelete, "/", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "errBody",
			req: httptest.NewRequest(http.MethodPost, "/", func() io.Reader {
				rcMock := NewReadCloser(t)
				rcMock.On("Read", mock.Anything).Return(0, errors.New("errBody"))
				return rcMock
			}()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "HdlrError",
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test")),
			mFunc: func(m *MockHandlers) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).Return(&url.URL{}, http.StatusConflict, errors.New("HdlrError"))
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "Correct",
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test")),
			mFunc: func(m *MockHandlers) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).Return(&url.URL{}, http.StatusOK, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hMock := NewMockHandlers(t)
			if tc.mFunc != nil {
				tc.mFunc(hMock)
			}

			s := &Server{handlers: hMock}
			w := httptest.NewRecorder()
			s.createShortURL(w, tc.req)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestUserURLs(t *testing.T) {
	ctx := context.WithValue(t.Context(), model.CtxUserID, "test")
	testCases := []struct {
		name       string
		req        *http.Request
		mFunc      func(m *MockHandlers)
		wantStatus int
	}{
		{
			name:       "WrongMethod",
			req:        httptest.NewRequest(http.MethodDelete, "/", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "UnAuth",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "HdlrError",
			req:  httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx),
			mFunc: func(m *MockHandlers) {
				m.On("GetUsersURLs", mock.Anything).Return(nil, errors.New("HdlrError"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "NoContent",
			req:  httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx),
			mFunc: func(m *MockHandlers) {
				m.On("GetUsersURLs", mock.Anything).Return(nil, nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "Correct",
			req:  httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx),
			mFunc: func(m *MockHandlers) {
				m.On("GetUsersURLs", mock.Anything).Return([]byte{0}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hMock := NewMockHandlers(t)
			if tc.mFunc != nil {
				tc.mFunc(hMock)
			}

			s := &Server{handlers: hMock}
			w := httptest.NewRecorder()
			s.userURLs(w, tc.req)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
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

func TestPingDB(t *testing.T) {
	testCases := []struct {
		name       string
		srv        *Server
		wantStatus int
	}{
		{
			name:       "NoDB",
			srv:        &Server{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "errPing",
			srv: &Server{db: func() handler.Database {
				dbMock := NewDatabase(t)
				dbMock.On("Ping", mock.Anything).Return(errors.New("dbErr"))
				return dbMock
			}()},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Correct",
			srv: &Server{db: func() handler.Database {
				dbMock := NewDatabase(t)
				dbMock.On("Ping", mock.Anything).Return(nil)
				return dbMock
			}()},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/", nil)
			tc.srv.pingDB(w, r)
			require.Equal(t, tc.wantStatus, w.Code)
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
	s, err := New(addr.Host, addr.String(), "", false, tmpStorage, nil)
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
