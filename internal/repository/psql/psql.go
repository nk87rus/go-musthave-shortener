package psql

import (
	"context"
	"iter"

	"github.com/jackc/pgx/v5"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2 --name=PSQLDriver --inpackage --testonly
type PSQLDriver interface {
	GetConnConfig() *pgx.ConnConfig
}

type Storage struct {
	db PSQLDriver
}

func NewStorage(ctx context.Context, dbDrv PSQLDriver) (*Storage, error) {
	if err := applyMigrations(ctx, dbDrv.GetConnConfig()); err != nil {
		return nil, err
	}
	return &Storage{db: dbDrv}, nil
}

func (s *Storage) Add(ctx context.Context, sURL, oURL string) error {
	return nil
}

func (s *Storage) Get(ctx context.Context, sURL string) (string, error) {
	return "", nil
}

func (s *Storage) IDExists(ctx context.Context, sURL string) bool {
	return false
}

func (s *Storage) LoadData(any) error {
	return nil
}

func (s *Storage) SaveData(iter.Seq[model.StorageRecord]) error {
	return nil
}
