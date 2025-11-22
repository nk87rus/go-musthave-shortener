package psql

import (
	"context"
	"encoding/json"
	"iter"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2 --name=PSQLDriver --inpackage --testonly
type PSQLDriver interface {
	GetConnConfig() *pgx.ConnConfig
	Insert(ctx context.Context, req string, args ...any) error
	InsertBatch(ctx context.Context, req string, args []pgx.NamedArgs) error
	SelectBytes(ctx context.Context, req string, args ...any) ([]byte, error)
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

func (s *Storage) Add(ctx context.Context, id, sURL, oURL string) error {
	req := `INSERT INTO public.urls(uuid, short_url, original_url) VALUES ($1, $2, $3);`
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	return s.db.Insert(ctx, req, id, sURL, oURL)
}

func (s *Storage) AddBatch(ctx context.Context, data iter.Seq[model.StorageRecord]) error {
	req := `INSERT INTO public.urls(uuid, short_url, original_url) VALUES (@uuidValue, @shortURL, @origURL);`
	var args = []pgx.NamedArgs{}

	for rec := range data {
		args = append(args, pgx.NamedArgs{"uuidValue": rec.UUID, "shortURL": rec.ShortURL, "origURL": rec.OrigURL})
	}

	ctx, cancelFunc := context.WithTimeout(ctx, reqTimeout(len(args)))
	defer cancelFunc()
	return s.db.InsertBatch(ctx, req, args)
}

func reqTimeout(value int) time.Duration {
	if value <= 5 {
		return 5 * time.Second
	}
	return time.Duration(value+value/2) * time.Second
}

// func (s *Storage) Get(ctx context.Context, sURL string) (string, error) {
// 	return "", nil
// }

// func (s *Storage) IDExists(ctx context.Context, sURL string) bool {
// 	return false
// }

func (s *Storage) LoadData(ctx context.Context, rcv any) error {
	req := `SELECT json_agg(row_to_json(r)) as data FROM (SELECT * FROM public.urls ORDER BY uuid ASC ) r`
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	rawData, err := s.db.SelectBytes(ctx, req)
	if err != nil {
		return err
	}

	if len(rawData) == 0 {
		rawData = []byte("[]")
	}

	if err := json.Unmarshal(rawData, rcv); err != nil {
		return err
	}

	return nil
}
