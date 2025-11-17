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
