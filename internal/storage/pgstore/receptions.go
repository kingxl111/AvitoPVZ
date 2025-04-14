package pgstore

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"AvitoPVZ/internal/infra/postgres"
	"AvitoPVZ/internal/receipt"
)

var receptionsTable = "receptions"

var receiptRowFields = fields(receiptRow{})

type receiptRow struct {
	ID            uuid.UUID `db:"id"`
	PvzID         uuid.UUID `db:"pvz_id"`
	ReceptionDate time.Time `db:"reception_date"`
	Status        string    `db:"status"`
}

func (rr receiptRow) ToModelReceipt() receipt.Receipt {
	return receipt.Receipt{
		ID:       rr.ID,
		PvzID:    rr.PvzID,
		DateTime: rr.ReceptionDate,
		Status:   receipt.Status(rr.Status),
	}
}

func (s *storageImpl) InsertReceipt(ctx context.Context, req receipt.InsertReceiptQuery) (*receipt.Receipt, error) {
	params := map[string]interface{}{
		"pvz_id":         req.PvzID,
		"reception_date": req.DateTime,
		"status":         req.Status,
	}

	q, args, err := s.stmpBuilder().
		Insert(receptionsTable).
		SetMap(params).
		Suffix("RETURNING " + receiptRowFields).
		ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	defer rows.Close()

	rr, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[receiptRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	m := rr.ToModelReceipt()
	return &m, nil
}

func (s *storageImpl) UpdateLastReceipt(ctx context.Context, req receipt.UpdateLastReceiptQuery) (*receipt.Receipt, error) {
	subQuery := s.stmpBuilder().
		Select("id").
		From(receptionsTable).
		Where(sq.Eq{
			"pvz_id": req.PvzID,
			"status": receipt.StatusInprogress,
		}).
		OrderBy("reception_date DESC").
		Limit(1)

	updateBuilder := s.stmpBuilder().
		Update(receptionsTable).
		Set("status", receipt.StatusClose).
		Set("reception_date", s.now()).
		Where(sq.Eq{"id": subQuery}).
		Suffix("RETURNING " + receiptRowFields)

	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	var r receiptRow
	err = s.db.QueryRow(ctx, sql, args...).Scan(&r.ID, &r.PvzID, &r.ReceptionDate, &r.Status)
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rec := r.ToModelReceipt()
	return &rec, nil
}
