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
	"github.com/nk87rus/go-musthave-shortener/internal/router/grpcsrv"
	"github.com/nk87rus/go-musthave-shortener/internal/router/httpsrv"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type SrvDatabase interface {
	handler.Database
	Close(ctx context.Context) error
}

//generate:reset
type App struct {
	httpServer *httpsrv.Server
	grpcServer *grpcsrv.Server
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

	if err := newApp.InitGRPCServer(cfg, storage); err != nil {
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
	newHTTPSrv, err := httpsrv.New(cfg.Addr, cfg.BaseAddr, cfg.TrustedSubnet, cfg.EnableTLS, storage, a.db)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg.AuditFile) != "" || strings.TrimSpace(cfg.AuditURL) != "" {
		newHTTPSrv.EnableAudit(strings.TrimSpace(cfg.AuditFile), strings.TrimSpace(cfg.AuditURL))
	}

	a.httpServer = newHTTPSrv
	return nil
}

func (a *App) InitGRPCServer(cfg *config.ConfigData, storage handler.Storage) error {
	newGRPCSrv, err := grpcsrv.New(cfg.GAddr, cfg.BaseAddr, storage)
	if err != nil {
		return err
	}

	a.grpcServer = newGRPCSrv
	return nil
}

func (a *App) Run(ctx context.Context) {
	defer func() {
		if a.db != nil {
			if err := a.db.Close(ctx); err != nil {
				log.Err(err)
			}
		}
	}()

	errGrp, egCtx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		return a.httpServer.Run(egCtx)
	})

	errGrp.Go(func() error {
		return a.grpcServer.Run(egCtx)
	})

	if err := errGrp.Wait(); err != nil {
		log.Fatal().Err(err)
	}
}
