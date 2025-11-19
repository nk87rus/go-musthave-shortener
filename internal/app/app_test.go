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
	"github.com/nk87rus/go-musthave-shortener/internal/repository/filestorage"
	memstorage "github.com/nk87rus/go-musthave-shortener/internal/repository/mem"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/psql"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	var (
		errConfig     = fmt.Errorf("errConfig")
		errExtStorage = fmt.Errorf("errExtStorage")
		errHTTP       = fmt.Errorf("errHTTP")
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
			name:      "errExtStorage",
			wantError: errExtStorage,
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

			patchStoreInit := monkey.Patch(memstorage.NewStorage,
				func(memstorage.ExtStorage) (*memstorage.MemStorage, error) {
					return &memstorage.MemStorage{}, nil
				})
			defer patchStoreInit.Unpatch()

			patchInitStorage := monkey.PatchInstanceMethod(reflect.TypeOf(&App{}), "InitExtStorage",
				func(*App, context.Context, *config.ConfigData) (memstorage.ExtStorage, error) {
					if errors.Is(tc.wantError, errExtStorage) {
						return nil, tc.wantError
					}
					return nil, nil
				})
			defer patchInitStorage.Unpatch()

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

func TestInitExtStorage(t *testing.T) {
	var (
		errNoStorage = fmt.Errorf("ошибка при инициализации storage")
		errDB        = fmt.Errorf("errDB")
		errDBStorage = fmt.Errorf("errDBStorage")
		errFS        = fmt.Errorf("errFS")
	)
	testCases := []struct {
		name      string
		cfg       config.ConfigData
		wantError error
	}{
		{
			name:      "errNoStorage",
			wantError: errNoStorage,
		},
		{
			name:      "errDB",
			cfg:       config.ConfigData{DBDSN: "db_test"},
			wantError: errDB,
		},
		{
			name:      "errDBStorage",
			cfg:       config.ConfigData{DBDSN: "db_test"},
			wantError: errDBStorage,
		},
		// {
		// 	name:      "errFS",
		// 	cfg:       config.ConfigData{FileStorage: "fs_test"},
		// 	wantError: errFS,
		// },
		{
			name:      "CorrectFS",
			cfg:       config.ConfigData{FileStorage: "fs_test"},
			wantError: nil,
		},
		{
			name:      "CorrectPQSL",
			cfg:       config.ConfigData{DBDSN: "db_test"},
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

			patchPSQLNewStorage := monkey.Patch(psql.NewStorage,
				func(context.Context, psql.PSQLDriver) (*psql.Storage, error) {
					if errors.Is(tc.wantError, errDBStorage) {
						return nil, tc.wantError
					}
					return new(psql.Storage), nil
				})
			defer patchPSQLNewStorage.Unpatch()

			patchFSt := monkey.Patch(filestorage.NewStorage,
				func(string) (*filestorage.Storage, error) {
					if errors.Is(tc.wantError, errFS) {
						return nil, tc.wantError
					}
					return new(filestorage.Storage), nil
				})
			// defer patchFSt.Unpatch()

			a := App{}
			resultData, resultError := a.InitExtStorage(context.Background(), &tc.cfg)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.NotNil(t, resultData)
			}
			patchFSt.Unpatch()
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
