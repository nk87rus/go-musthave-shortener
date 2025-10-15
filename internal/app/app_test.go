package app

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	patchStoreInit := monkey.Patch(simple.NewStorage,
		func() *simple.Storage {
			return &simple.Storage{}
		})
	defer patchStoreInit.Unpatch()

	patchNewHttp := monkey.Patch(httpsrv.New,
		func(handler.Storage) *httpsrv.Server {
			return &httpsrv.Server{}
		})
	defer patchNewHttp.Unpatch()

	require.IsType(t, &App{}, Init())
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
			patchHttpSrvRun := monkey.PatchInstanceMethod(reflect.TypeOf(&httpsrv.Server{}), "Run",
				func(*httpsrv.Server) error {
					if errors.Is(tc.wantError, errRun) {
						return tc.wantError
					}
					return nil
				})
			defer patchHttpSrvRun.Unpatch()

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
