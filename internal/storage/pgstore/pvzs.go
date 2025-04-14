package pgstore

import (
	"context"
	"time"

	"AvitoPVZ/internal/infra/postgres"
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	pvzsTable    = "pvzs"
	defaultPage  = 1
	defaultLimit = 50
)

var pvzRowFields = fields(pvzRow{})

type pvzRow struct {
	ID             uuid.UUID `db:"id"`
	City           string    `db:"city"`
	DateRegistered time.Time `db:"date_registered"`
}

func (pr pvzRow) ToModel() *pvz.PVZ {
	return &pvz.PVZ{
		ID:               pr.ID,
		City:             pvz.City(pr.City),
		RegistrationDate: pr.DateRegistered.Format(time.RFC3339),
	}
}

func (s *storageImpl) InsertPVZ(ctx context.Context, req pvz.InsertPVZQuery) (*pvz.PVZ, error) {
	params := map[string]interface{}{
		"city": req.City,
	}
	if req.RegistrationDate != nil {
		params["date_registered"] = *req.RegistrationDate
	} else {
		params["date_registered"] = s.now()
	}
	if req.ID != nil {
		params["id"] = *req.ID
	}

	q, args, err := s.stmpBuilder().
		Insert(pvzsTable).
		SetMap(params).
		Suffix("RETURNING " + pvzRowFields).
		ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	defer rows.Close()

	pr, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[pvzRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	return pr.ToModel(), nil
}

func (s *storageImpl) ListPVZs(ctx context.Context, req pvz.ListPVZsQuery) ([]pvz.PVZ, error) {
	builder := s.stmpBuilder().
		Select("DISTINCT " + pvzRowFields).
		From(pvzsTable).
		Join("receptions ON pvzs.id = receptions.pvz_id")

	if req.StartDate != nil || req.EndDate != nil {
		builder = builder.Where(sq.NotEq{"receptions.reception_date": nil})
	}

	if req.StartDate != nil {
		startTime, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			return nil, postgres.HandleError(err)
		}
		builder = builder.Where(sq.GtOrEq{"receptions.reception_date": startTime})
	}
	if req.EndDate != nil {
		endTime, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return nil, postgres.HandleError(err)
		}
		builder = builder.Where(sq.LtOrEq{"receptions.reception_date": endTime})
	}

	page := defaultPage
	limit := defaultLimit
	if req.Page != nil {
		page = *req.Page
	}
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := (page - 1) * limit
	builder = builder.Offset(uint64(offset)).Limit(uint64(limit)).
		OrderBy("pvzs.date_registered DESC")

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	defer rows.Close()

	pvzRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[pvzRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	var pvzs []pvz.PVZ
	for _, pr := range pvzRows {
		pvzs = append(pvzs, *pr.ToModel())
	}

	return pvzs, nil
}

func (s *storageImpl) ListReceptions(ctx context.Context, req receipt.ListReceptionsQuery) ([]receipt.Receipt, error) {
	builder := s.stmpBuilder().
		Select(receiptRowFields).
		From(receptionsTable).
		Where(sq.Eq{"pvz_id": req.PvzIDs})

	if req.StartDate != nil {
		startTime, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			return nil, postgres.HandleError(err)
		}
		builder = builder.Where(sq.GtOrEq{"reception_date": startTime})
	}
	if req.EndDate != nil {
		endTime, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return nil, postgres.HandleError(err)
		}
		builder = builder.Where(sq.LtOrEq{"reception_date": endTime})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	defer rows.Close()

	receiptRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[receiptRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	var receipts []receipt.Receipt
	for _, rr := range receiptRows {
		receipts = append(receipts, rr.ToModelReceipt())
	}

	return receipts, nil
}
