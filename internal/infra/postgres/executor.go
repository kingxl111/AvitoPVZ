package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type (
	Executor interface {
		Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
		Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
		QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	}
)

type Tx pgx.Tx
