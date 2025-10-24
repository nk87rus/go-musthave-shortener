package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/app"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	var errInit = fmt.Errorf("errInit")

	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errInit",
			wantError: errInit,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	patchExit := monkey.Patch(os.Exit,
		func(int) {
			panic(errInit)
		})
	defer patchExit.Unpatch()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchAppInit := monkey.Patch(app.Init,
				func() (*app.App, error) {
					if errors.Is(tc.wantError, errInit) {
						return nil, tc.wantError
					}
					return &app.App{}, nil
				})
			defer patchAppInit.Unpatch()

			patchAppRun := monkey.PatchInstanceMethod(reflect.TypeOf(&app.App{}), "Run",
				func(*app.App) {
					//nolint:funlen
				})
			defer patchAppRun.Unpatch()

			if tc.wantError != nil {
				require.PanicsWithError(t, tc.wantError.Error(), main)
			} else {
				main()
			}
		})
	}
}
