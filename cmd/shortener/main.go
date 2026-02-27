package main

import (
	"context"
	"os"

	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

func main() {
	ctx := context.Background()
	app, err := app.Init(ctx)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	go app.Run(ctx)
	http.ListenAndServe(":7080", nil)
}
