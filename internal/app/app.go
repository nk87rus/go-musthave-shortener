package app

import (
	"context"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/config/db"
	"github.com/nk87rus/go-musthave-shortener/internal/logger"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/rs/zerolog/log"
)
//go:generate go run github.com/vektra/mockery/v2 --name=SrvDatabase --inpackage --testonly
type SrvDatabase interface {
	Close(ctx context.Context) error
}

type App struct {
	httpServer *httpsrv.Server
	db         SrvDatabase
}

func Init(ctx context.Context) (*App, error) {
	logger.Init()
	cfg, err := config.InitConfig(os.Args)
	if err != nil {
		return nil, err
	}

	log.Info().Any("cfg", cfg).Msg("Сфоромирована конфигурация")

	storage, err := simple.NewStorage(cfg.FileStorage)
	if err != nil {
		return nil, err
	}

	psql, err := db.InitPSQL(ctx, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, storage, psql)
	if err != nil {
		return nil, err
	}
	return &App{httpServer: newHTTPSrv, db: psql}, nil
}

func (a *App) Run(ctx context.Context) {
	defer func() {
		if a.db != nil {
			a.db.Close(ctx)
		}
	}()

	if err := a.httpServer.Run(ctx); err != nil {
		log.Fatal().Err(err)
	}
}
