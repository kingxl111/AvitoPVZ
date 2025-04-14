package story

import (
	"AvitoPVZ/internal/product"
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"AvitoPVZ/internal/storage"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (s *Story) AddProduct(ctx context.Context, req product.AddProductRequest) (*product.Product, error) {
	var inserted *product.Product
	err := s.txManager.Tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rcpt, err := s.storage.GetLastReceipt(ctx, receipt.GetLastReceiptQuery{
			PvzID:     req.PvzID,
			ForUpdate: true,
		})
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return pvz.ErrPVZNotFound
			}
		}

		if rcpt.Status == receipt.StatusClose {
			return receipt.ErrLastReceiptClosed
		}
		inserted, err = s.storage.InsertProduct(ctx, product.InsertProductQuery{
			Type:  req.Type,
			PvzID: req.PvzID,
		})
		return err
	}, nil)
	if err != nil {
		return nil, err
	}

	return inserted, nil
}
