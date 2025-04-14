package story

import (
	"AvitoPVZ/internal/product"
	"AvitoPVZ/internal/receipt"
	"context"

	"github.com/jackc/pgx/v5"
)

type (
	stg interface {
		InsertProduct(ctx context.Context, req product.InsertProductQuery) (*product.Product, error)
		DeleteLastProduct(ctx context.Context, req product.DeleteLastProductQuery) (int64, error)
		GetLastReceipt(ctx context.Context, req receipt.GetLastReceiptQuery) (*receipt.Receipt, error)
	}
	manager interface {
		Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) (err error)
	}
)
