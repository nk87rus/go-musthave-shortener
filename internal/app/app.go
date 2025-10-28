package app

import (
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/logger"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/rs/zerolog/log"
)

type App struct {
	httpServer *httpsrv.Server
}

func Init() (*App, error) {
	logger.Init()
	cfg, err := config.InitConfig(os.Args)
	if err != nil {
		return nil, err
	}

	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, simple.NewStorage())
	if err != nil {
		return nil, err
	}
	return &App{httpServer: newHTTPSrv}, nil
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		log.Fatal().Err(err)
	}
}
