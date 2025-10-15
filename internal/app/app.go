package app

import (
	"log"

	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
)

type App struct {
	httpServer *httpsrv.Server
}

func Init() *App {
	store := simple.NewStorage()
	return &App{
		httpServer: httpsrv.New(store),
	}
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
