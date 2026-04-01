package grpcsrv

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	pb "github.com/nk87rus/go-musthave-shortener/internal/router/grpcsrv/proto"
	"github.com/stretchr/testify/mock"
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
				handlers: new(handler.Handlers),
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultData, resultError := New(tc.addr, tc.baddr, nil)
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

func TestShortenURL(t *testing.T) {
	var (
		errAuth = fmt.Errorf("errAuth")
		errHDLR = fmt.Errorf("errHDLR")
	)

	testCases := []struct {
		name      string
		mFunc     func(m *MockHandlers)
		wantError error
	}{
		{
			name:      "errAuth",
			wantError: errAuth,
		},
		{
			name: "errHDLR",
			mFunc: func(m *MockHandlers) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).Return(nil, 0, errHDLR)
			},
			wantError: errHDLR,
		},
		{
			name: "Correct",
			mFunc: func(m *MockHandlers) {
				m.On("CreateShortURL", mock.Anything, mock.Anything).Return(&url.URL{RawPath: "http://test"}, 1, nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchGAD := monkey.Patch(getAuthData,
				func(context.Context) (string, error) {
					if errors.Is(tc.wantError, errAuth) {
						return "", tc.wantError
					}
					return "test", nil
				})
			defer patchGAD.Unpatch()

			hMock := NewMockHandlers(t)
			if tc.mFunc != nil {
				tc.mFunc(hMock)
			}

			gs := &Server{handlers: hMock}
			resultData, resultError := gs.ShortenURL(t.Context(), &pb.URLShortenRequest{})
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Empty(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, resultData)
			}
		})
	}
}
