package app

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	var (
		errConfig = fmt.Errorf("errConfig")
		errHTTP   = fmt.Errorf("errHTTP")
	)
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errConfig",
			wantError: errConfig,
		},
		{
			name:      "errHTTP",
			wantError: errHTTP,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchInitConfig := monkey.Patch(config.InitConfig,
				func([]string) (*config.ConfigData, error) {
					if errors.Is(tc.wantError, errConfig) {
						return nil, tc.wantError
					}
					return &config.ConfigData{}, nil
				})
			defer patchInitConfig.Unpatch()

			patchStoreInit := monkey.Patch(simple.NewStorage,
				func() *simple.Storage {
					return &simple.Storage{}
				})
			defer patchStoreInit.Unpatch()

			patchNewHTTP := monkey.Patch(httpsrv.New,
				func(string, string, handler.Storage) (*httpsrv.Server, error) {
					if errors.Is(tc.wantError, errHTTP) {
						return nil, tc.wantError
					}
					return &httpsrv.Server{}, nil
				})
			defer patchNewHTTP.Unpatch()

			resultData, resultError := Init()
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &App{}, resultData)
			}
		})
	}
}

func TestAppRun(t *testing.T) {
	errRun := fmt.Errorf("errRun")

	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errRun",
			wantError: errRun,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchHTTPSrvRun := monkey.PatchInstanceMethod(reflect.TypeOf(&httpsrv.Server{}), "Run",
				func(*httpsrv.Server) error {
					if errors.Is(tc.wantError, errRun) {
						return tc.wantError
					}
					return nil
				})
			defer patchHTTPSrvRun.Unpatch()

			var catchedErr error
			patchLogFatal := monkey.Patch(log.Fatal,
				func(v ...any) {
					switch cv := v[0].(type) {
					case error:
						catchedErr = cv
					default:
						t.Fatal("не корретный тип параметра")
					}
				})
			defer patchLogFatal.Unpatch()
			a := &App{}
			a.Run()
			if tc.wantError != nil {
				require.ErrorIs(t, errRun, catchedErr)
			}
		})
	}
}
