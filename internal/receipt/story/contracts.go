package story

import (
	"AvitoPVZ/internal/receipt"
	"context"

	"github.com/jackc/pgx/v5"
)

type (
	stg interface {
		InsertReceipt(ctx context.Context, req receipt.InsertReceiptQuery) (*receipt.Receipt, error)
		UpdateLastReceipt(ctx context.Context, req receipt.UpdateLastReceiptQuery) (*receipt.Receipt, error)
		GetLastReceipt(ctx context.Context, req receipt.GetLastReceiptQuery) (*receipt.Receipt, error)
	}
	manager interface {
		Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) (err error)
	}
)
