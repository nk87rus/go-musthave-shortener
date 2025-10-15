package main

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

func TestMain(t *testing.T) {
	patchAppInit := monkey.Patch(app.Init,
		func() *app.App {
			return &app.App{}
		})
	defer patchAppInit.Unpatch()

	patchAppRun := monkey.PatchInstanceMethod(reflect.TypeOf(&app.App{}), "Run",
		func(*app.App) {
			//nolint:funlen
		})
	defer patchAppRun.Unpatch()
	main()
}
