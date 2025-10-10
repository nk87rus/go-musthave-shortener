package app

import (
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
	a.httpServer.Run()
}
