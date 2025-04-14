package story

import (
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"AvitoPVZ/internal/storage"
	"context"

	"github.com/pkg/errors"
)

func (s *Story) CloseLast(ctx context.Context, req receipt.CloseLastReceiptRequest) (*receipt.Receipt, error) {
	updatedReceipt, err := s.storage.UpdateLastReceipt(ctx, receipt.UpdateLastReceiptQuery{
		PvzID:  req.PvzID,
		Status: receipt.StatusClose,
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, pvz.ErrPVZNotFound
		}
		return nil, err
	}

	return updatedReceipt, nil
}
