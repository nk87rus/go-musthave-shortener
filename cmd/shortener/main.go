package main

import (
	"context"

	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

func main() {
	ctx := context.Background()
	app, err := app.Init(ctx)
	if err != nil {
		println(err.Error())
		return
	}
	go app.Run(ctx)
	if errHTTP := http.ListenAndServe(":7080", nil); errHTTP != nil {
		println(errHTTP.Error())
	}
}
