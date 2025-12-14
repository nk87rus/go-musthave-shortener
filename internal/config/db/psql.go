package db

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

const PSQLDSN = "postgres://postgres:1234@localhost:5432/shortener"

type PSQL struct {
	conn *pgx.Conn
}

func InitPSQL(ctx context.Context, connString string) (*PSQL, error) {
	log.Info().Str("connString", connString).Msg("Инициализация подключения к PSQL")
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PSQL{conn: conn}, nil
}

func (p *PSQL) Close(ctx context.Context) error {
	return p.conn.Close(ctx)
}

func (p *PSQL) Ping(ctx context.Context) error {
	err := p.conn.Ping(ctx)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return errors.New(strings.Trim(pgErr.Message, "\n"))
	}
	return err
}

func (p *PSQL) GetConnConfig() *pgx.ConnConfig {
	return p.conn.Config()
}

func (p *PSQL) Insert(ctx context.Context, req string, args ...any) error {
	_, err := p.conn.Exec(ctx, req, args...)
	if err != nil {
		return err
	}
	return nil
}

func (p *PSQL) InsertBatch(ctx context.Context, req string, args []pgx.NamedArgs) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, a := range args {
		batch.Queue(req, a)

		if batch.Len() == 1000 {
			if err := sendBatch(ctx, tx, batch); err != nil {
				return errors.Join(err, tx.Rollback(ctx))
			}
			batch = &pgx.Batch{}
		}
	}

	if batch.Len() > 0 {
		if err := sendBatch(ctx, tx, batch); err != nil {
			return errors.Join(err, tx.Rollback(ctx))
		}
	}

	tx.Commit(ctx)
	return nil
}

func sendBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch) error {
	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	if _, err := results.Exec(); err != nil {
		return err
	}

	return nil
}

func (p *PSQL) SelectBytes(ctx context.Context, req string, args ...any) ([]byte, error) {
	return dbSelect[[]byte](ctx, p.conn, req, args...)
}

func (p *PSQL) SelectString(ctx context.Context, req string, args ...any) (string, error) {
	return dbSelect[string](ctx, p.conn, req, args...)
}

func dbSelect[T []byte | string](ctx context.Context, cli *pgx.Conn, req string, args ...any) (T, error) {
	var dbResponse T
	err := cli.QueryRow(ctx, req, args...).Scan(&dbResponse)
	return dbResponse, err
}

func (p *PSQL) Exec(ctx context.Context, req string, args ...any) error {
	_, err := p.conn.Exec(ctx, req, args...)
	return err
}
