package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/nk87rus/go-musthave-shortener/internal/repository"
)

//go:generate go run github.com/vektra/mockery/v2 --name=PSQLDriver --inpackage --testonly
type PSQLDriver interface {
	GetConnConfig() *pgx.ConnConfig
	Insert(ctx context.Context, req string, args ...any) error
	InsertBatch(ctx context.Context, req string, args []pgx.NamedArgs) error
	SelectBytes(ctx context.Context, req string, args ...any) ([]byte, error)
	SelectString(ctx context.Context, req string, args ...any) (string, error)
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
	userID, ok := ctx.Value(model.CtxUserID).(string)
	if !ok {
		return fmt.Errorf("не корректный тип userID (%T)", ctx.Value(model.CtxUserID))
	}

	req := `INSERT INTO public.urls(uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4);`
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	if err := s.db.Insert(ctx, req, id, sURL, oURL, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
				defer cancelFunc()
				curShortURL, sURLErr := s.GetSortURL(ctx, oURL)
				if sURLErr != nil {
					return err
				}
				return &repository.DBError{Err: err, HTTPResponseCode: http.StatusConflict, Value: curShortURL}
			}
		}
		return fmt.Errorf("Add: %w", err)
	}
	return nil
}

func (s *Storage) AddBatch(ctx context.Context, data iter.Seq[model.StorageRecord]) error {
	req := `INSERT INTO public.urls(uuid, short_url, original_url, user_id) VALUES (@uuidValue, @shortURL, @origURL, @userID);`
	var args = []pgx.NamedArgs{}

	for rec := range data {
		args = append(args, pgx.NamedArgs{"uuidValue": rec.UUID, "shortURL": rec.ShortURL, "origURL": rec.OrigURL, "userID": rec.UserID})
	}

	ctx, cancelFunc := context.WithTimeout(ctx, reqTimeout(len(args)))
	defer cancelFunc()
	if err := s.db.InsertBatch(ctx, req, args); err != nil {
		return fmt.Errorf("AddBatch: %w", err)
	}
	return nil
}

func reqTimeout(value int) time.Duration {
	if value <= 5 {
		return 5 * time.Second
	}
	return time.Duration(value+value/2) * time.Second
}

func (s *Storage) GetSortURL(ctx context.Context, origURL string) (string, error) {
	req := "SELECT short_url FROM public.urls WHERE original_url = $1;"
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	result, err := s.db.SelectString(ctx, req, origURL)
	if err != nil {
		return "", fmt.Errorf("GetSortURL: %w", err)
	}
	return result, nil
}

// func (s *Storage) IDExists(ctx context.Context, sURL string) bool {
// 	return false
// }

func (s *Storage) LoadData(ctx context.Context, rcv any) error {
	req := `SELECT json_agg(row_to_json(r)) as data FROM (SELECT * FROM public.urls ORDER BY uuid ASC ) r`
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	rawData, err := s.db.SelectBytes(ctx, req)
	if err != nil {
		return fmt.Errorf("LoadData: %w", err)
	}

	if len(rawData) == 0 {
		rawData = []byte("[]")
	}

	if err := json.Unmarshal(rawData, rcv); err != nil {
		return fmt.Errorf("LoadData: %w", err)
	}

	return nil
}
