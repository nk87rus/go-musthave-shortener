//go:generate go run github.com/vektra/mockery/v2 --all --inpackage --testonly
package app

import (
	"context"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/config/db"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/logger"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/simple"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/rs/zerolog/log"
)

type SrvDatabase interface {
	handler.Database
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

	var newApp = App{}

	storage, err := simple.NewStorage(cfg.FileStorage)
	if err != nil {
		return nil, err
	}

	if cfg.DBDSN != "" {
		if err := newApp.InitDBConnection(ctx, cfg.DBDSN); err != nil {
			return nil, err
		}
	}

	if err := newApp.InitHTTPServer(cfg.Addr, cfg.BaseAddr, storage); err != nil {
		return nil, err
	}

	return &newApp, nil
}

func (a *App) InitDBConnection(ctx context.Context, dsn string) error {
	psql, err := db.InitPSQL(ctx, dsn)
	if err != nil {
		return err
	}
	a.db = psql
	return nil
}

func (a *App) InitHTTPServer(addr, baseAddr string, storage handler.Storage) error {
	newHTTPSrv, err := httpsrv.New(addr, baseAddr, storage, a.db)
	if err != nil {
		return err
	}
	a.httpServer = newHTTPSrv
	return nil
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
