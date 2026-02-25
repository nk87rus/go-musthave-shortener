package httpsrv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nk87rus/go-musthave-shortener/internal/handler"
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
	go s.Run(context.Background())
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
