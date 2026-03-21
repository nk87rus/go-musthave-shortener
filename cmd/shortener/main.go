package main

import (
	"context"
	"fmt"
	"strings"

	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/nk87rus/go-musthave-shortener/internal/app"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", getValue(buildVersion), getValue(buildDate), getValue(buildCommit))
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

func showBuildData() {
}

func getValue(data string) string {
	if strings.TrimSpace(data) == "" {
		return "N/A"
	}
	return data
}
