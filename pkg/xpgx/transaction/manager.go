package transaction

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"AvitoPVZ/pkg/xpgx"
)

type Manager interface {
	Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) error
}

type managerImpl struct {
	db *xpgx.DB
}

func NewManager(db *xpgx.DB) Manager {
	return &managerImpl{
		db: db,
	}
}

func (m *managerImpl) Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) (err error) {
	db := m.db

	var tx pgx.Tx
	if opts == nil {
		tx, err = db.Begin(ctx)
	} else {
		tx, err = db.BeginTx(ctx, *opts)
	}
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tryRollback(ctx, tx)
			panic(p)
		}
	}()

	err = fn(ctx, tx)
	if err != nil {
		tryRollback(ctx, tx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return err
}

func tryRollback(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "StorageManager.Transaction: tx.Rollback: ", slog.Any("err", err))
	}
}
