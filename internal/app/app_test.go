package app

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/config/db"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	var (
		errConfig = fmt.Errorf("errConfig")
		errDB     = fmt.Errorf("errDB")
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
			name:      "errDB",
			wantError: errDB,
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
					return &config.ConfigData{DBDSN: "DSN"}, nil
				})
			defer patchInitConfig.Unpatch()

			patchStoreInit := monkey.Patch(simple.NewStorage,
				func(string) (*simple.Storage, error) {
					return &simple.Storage{}, nil
				})
			defer patchStoreInit.Unpatch()

			patchInitDB := monkey.PatchInstanceMethod(reflect.TypeOf(&App{}), "InitDBConnection",
				func(*App, context.Context, string) error {
					if errors.Is(tc.wantError, errDB) {
						return tc.wantError
					}
					return nil
				})

			defer patchInitDB.Unpatch()
			patchInitHTTP := monkey.PatchInstanceMethod(reflect.TypeOf(&App{}), "InitHTTPServer",
				func(*App, string, string, handler.Storage) error {
					if errors.Is(tc.wantError, errHTTP) {
						return tc.wantError
					}
					return nil
				})
			defer patchInitHTTP.Unpatch()

			resultData, resultError := Init(context.Background())
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &App{}, resultData)
			}
		})
	}
}

func TestInitDBConnection(t *testing.T) {
	var errDB = fmt.Errorf("errDB")
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errDB",
			wantError: errDB,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchPSQL := monkey.Patch(db.InitPSQL,
				func(context.Context, string) (*db.PSQL, error) {
					if errors.Is(tc.wantError, errDB) {
						return nil, tc.wantError
					}
					return new(db.PSQL), nil
				})
			defer patchPSQL.Unpatch()

			a := App{}
			resultError := a.InitDBConnection(context.Background(), "Test DSN")
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, a.db)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &db.PSQL{}, a.db)
			}
		})
	}
}

func TestInitHTTPSrv(t *testing.T) {
	var errHTTP = fmt.Errorf("errHTTP")
	testCases := []struct {
		name      string
		wantError error
	}{
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
			patchNewHTTP := monkey.Patch(httpsrv.New,
				func(string, string, handler.Storage, handler.Database) (*httpsrv.Server, error) {
					if errors.Is(tc.wantError, errHTTP) {
						return nil, tc.wantError
					}
					return &httpsrv.Server{}, nil
				})
			defer patchNewHTTP.Unpatch()

			a := App{}
			resultError := a.InitHTTPServer("addr", "baddr", nil)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, a.httpServer)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &httpsrv.Server{}, a.httpServer)
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
				func(*httpsrv.Server, context.Context) error {
					if errors.Is(tc.wantError, errRun) {
						return tc.wantError
					}
					return nil
				})
			defer patchHTTPSrvRun.Unpatch()

			dbMock := NewMockSrvDatabase(t)
			dbMock.On("Close", mock.Anything).Return(nil)
			a := &App{db: dbMock}
			a.Run(context.Background())

		})
	}
}
