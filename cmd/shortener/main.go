package main

import (
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

func main() {
	app, err := app.Init()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	app.Run()
}
