package story

import (
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"AvitoPVZ/internal/storage"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (s *Story) Create(ctx context.Context, req receipt.CreateReceiptRequest) (*receipt.Receipt, error) {
	var insertedReceipt *receipt.Receipt
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

		if rcpt.Status != receipt.StatusClose {
			return receipt.ErrLastReceiptNotClosed
		}

		insertedReceipt, err = s.storage.InsertReceipt(ctx, receipt.InsertReceiptQuery{
			PvzID:    req.PvzID,
			DateTime: s.now(),
			Status:   receipt.StatusInprogress,
		})
		if err != nil {
			return errors.WithMessage(err, "storage.InsertReceipt")
		}
		return err
	}, nil)
	if err != nil {
		return nil, err
	}

	return insertedReceipt, nil
}
