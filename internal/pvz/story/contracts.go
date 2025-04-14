package story

import (
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"context"

	"github.com/jackc/pgx/v5"
)

type (
	stg interface {
		InsertPVZ(ctx context.Context, req pvz.InsertPVZQuery) (*pvz.PVZ, error)
		ListPVZs(ctx context.Context, req pvz.ListPVZsQuery) ([]pvz.PVZ, error)
		ListReceptions(ctx context.Context, req receipt.ListReceptionsQuery) ([]receipt.Receipt, error)
	}
	manager interface {
		Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) (err error)
	}
)
