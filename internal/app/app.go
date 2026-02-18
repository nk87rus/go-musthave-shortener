//go:generate go run github.com/vektra/mockery/v2 --all --inpackage --testonly
package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nk87rus/go-musthave-shortener/internal/config"
	"github.com/nk87rus/go-musthave-shortener/internal/config/db"
	"github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/logger"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/filestorage"
	memstorage "github.com/nk87rus/go-musthave-shortener/internal/repository/mem"
	"github.com/nk87rus/go-musthave-shortener/internal/repository/psql"
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

	extStorage, err := newApp.InitExtStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}

	storage, err := memstorage.NewStorage(ctx, extStorage)
	if err != nil {
		return nil, err
	}

	if err := newApp.InitHTTPServer(cfg, storage); err != nil {
		return nil, err
	}

	return &newApp, nil
}

func (a *App) InitExtStorage(ctx context.Context, cfg *config.ConfigData) (memstorage.ExtStorage, error) {
	switch {
	case cfg.DBDSN != "":
		psqlDrv, err := db.InitPSQL(ctx, cfg.DBDSN)
		if err != nil {
			return nil, err
		}
		a.db = psqlDrv

		pstr, err := psql.NewStorage(ctx, psqlDrv)
		if err != nil {
			return nil, err
		}
		return pstr, nil
	case cfg.FileStorage != "":
		newFS, err := filestorage.NewStorage(cfg.FileStorage)
		if err != nil {
			return nil, err
		}
		return newFS, nil
	}
	return nil, fmt.Errorf("ошибка при инициализации storage")
}

func (a *App) InitHTTPServer(cfg *config.ConfigData, storage handler.Storage) error {
	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, storage, a.db)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg.AuditFile) != "" || strings.TrimSpace(cfg.AuditURL) != "" {
		newHTTPSrv.EnableAudit(strings.TrimSpace(cfg.AuditFile), strings.TrimSpace(cfg.AuditURL))
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
