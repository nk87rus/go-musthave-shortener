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

func Init() (*App, error) {
	cfg := config.InitConfig()
	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, simple.NewStorage())
	if err != nil {
		return nil, err
	}
	return &App{httpServer: newHTTPSrv}, nil
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
