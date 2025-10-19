package app

import (
	"log"

	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
)

type App struct {
	httpServer *httpsrv.Server
}

func Init() *App {
	cfg := config.InitConfig()
	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, simple.NewStorage())
	if err != nil {
		log.Fatal(err)
	}
	return &App{httpServer: newHTTPSrv}
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
