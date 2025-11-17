package main

import (
	"context"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

func main() {
	ctx := context.Background()
	app, err := app.Init(ctx)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	app.Run(ctx)
}
