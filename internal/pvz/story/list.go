package story

import (
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func (s *Story) ListPVZWithReceipts(ctx context.Context, req pvz.ListPVZWithReceiptsRequest) ([]pvz.PVZWithReceipts, error) {
	var pwrs []pvz.PVZWithReceipts
	err := s.txManager.Tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		pvzs, err := s.storage.ListPVZs(ctx, pvz.ListPVZsQuery{
			Page:      req.Page,
			Limit:     req.Limit,
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
		})
		if err != nil {
			return errors.WithMessage(err, "storage.ListPVZs")
		}
		if len(pvzs) == 0 {
			return nil
		}

		receptions, err := s.storage.ListReceptions(ctx, receipt.ListReceptionsQuery{
			PvzIDs: lo.Map(pvzs, func(p pvz.PVZ, _ int) uuid.UUID {
				return p.ID
			}),
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
		})

		pwrs = lo.Map(pvzs, func(p pvz.PVZ, _ int) pvz.PVZWithReceipts {
			return pvz.PVZWithReceipts{
				PVZ: p,
				Receipts: lo.Filter(receptions, func(r receipt.Receipt, _ int) bool {
					return r.PvzID == p.ID
				}),
			}
		})

		return nil
	}, nil)
	if err != nil {
		return nil, err
	}

	return pwrs, nil
}
