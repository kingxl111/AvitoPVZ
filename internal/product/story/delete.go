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

func (s *Story) Delete(ctx context.Context, req product.DeleteLastProductRequest) (n int64, err error) {
	err = s.txManager.Tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
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

		n, err = s.storage.DeleteLastProduct(ctx, product.DeleteLastProductQuery{
			PvzID: req.PvzID,
		})
		return err
	}, nil)
	if err != nil {
		return 0, err
	}

	return n, nil
}
