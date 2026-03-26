package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	log.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", getValue(buildVersion), getValue(buildDate), getValue(buildCommit))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	app, err := app.Init(ctx)
	if err != nil {
		log.Fatalf("Ошибка при инициализации приложения: %v", err)
	}
	go app.Run(ctx)

	srv := http.Server{Addr: ":7080"}
	go func() {
		if errHTTP := srv.ListenAndServe(); errHTTP != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера профилирования: %v", errHTTP)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Ошибка graceful shutdown: %v", err)
	}

	log.Println("Graceful shutdown выполнен успешно")
}

func getValue(data string) string {
	if strings.TrimSpace(data) == "" {
		return "N/A"
	}
	return data
}
