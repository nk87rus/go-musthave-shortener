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

//go:generate go run github.com/vektra/mockery/v2 --name=DBConn --inpackage --testonly
type DBConn interface {
	Close(context.Context) error
	IsClosed() bool
	Ping(ctx context.Context) error
	Config() *pgx.ConnConfig
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PSQL struct {
	conn DBConn
}

func InitPSQL(ctx context.Context, connString string) (*PSQL, error) {
	log.Info().Str("connString", connString).Msg("Инициализация подключения к PSQL")
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	pConn := PSQL{conn: conn}
	go pConn.GracefulShutdown(ctx)

	return &pConn, nil
}

func (p *PSQL) GracefulShutdown(ctx context.Context) {
	<-ctx.Done()
	log.Debug().Str("packet", "db").Msg("graceful shutdown - start")
	if err := p.Close(ctx); err != nil {
		log.Err(err).Str("package", "db").Msg("graceful shutdown - failed")
		return
	}
	log.Debug().Str("packet", "db").Msg("graceful shutdown - success")
}

func (p *PSQL) Close(ctx context.Context) error {
	if !p.conn.IsClosed() {
		return p.conn.Close(ctx)
	}
	return nil
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

	return tx.Commit(ctx)
}

func sendBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch) error {
	results := tx.SendBatch(ctx, batch)
	defer func() {
		if err := results.Close(); err != nil {
			log.Err(err)
		}
	}()

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

func dbSelect[T []byte | string](ctx context.Context, cli DBConn, req string, args ...any) (T, error) {
	var dbResponse T
	err := cli.QueryRow(ctx, req, args...).Scan(&dbResponse)
	return dbResponse, err
}

func (p *PSQL) Exec(ctx context.Context, req string, args ...any) error {
	_, err := p.conn.Exec(ctx, req, args...)
	return err
}
