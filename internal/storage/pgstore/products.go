package pgstore

import (
	"AvitoPVZ/internal/infra/postgres"
	"AvitoPVZ/internal/product"
	"AvitoPVZ/internal/receipt"
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

const (
	productsTable = "products"
)

type productRow struct {
	ID          uuid.UUID `db:"id"`
	ReceptionID uuid.UUID `db:"reception_id"`
	AddedAt     time.Time `db:"added_at"`
	ProductType string    `db:"product_type"`
	PvzID       uuid.UUID `db:"pvz_id"`
}

func (p productRow) ToModel() product.Product {
	return product.Product{
		ID:       p.ID,
		PvzID:    p.PvzID,
		Type:     product.Type(p.ProductType),
		DateTime: p.AddedAt,
	}
}

func (s *storageImpl) InsertProduct(ctx context.Context, req product.InsertProductQuery) (*product.Product, error) {
	subQuery := s.stmpBuilder().
		Select("id").
		From(receptionsTable).
		Where(sq.Eq{
			"pvz_id": req.PvzID,
			"status": receipt.StatusInprogress,
		}).
		OrderBy("reception_date DESC").
		Limit(1)

	sql, args, err := subQuery.ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	var receptionID uuid.UUID
	err = s.db.QueryRow(ctx, sql, nil).Scan(&receptionID)
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	insertBuilder := s.stmpBuilder().
		Insert(productsTable).
		Columns("reception_id", "product_type", "added_at").
		Values(receptionID, req.Type, s.now()).
		Suffix("RETURNING " + fields(productRow{}))

	sql, args, err = insertBuilder.ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	var p productRow
	err = s.db.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.ReceptionID, &p.AddedAt, &p.ProductType)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	pModel := p.ToModel()

	return &pModel, nil
}

func (s *storageImpl) DeleteLastProduct(ctx context.Context, req product.DeleteLastProductQuery) (int64, error) {
	subQuery := s.stmpBuilder().
		Select("p.id").
		From(productsTable + " p").
		Join(receptionsTable + " r ON p.reception_id = r.id").
		Where(sq.Eq{
			"r.pvz_id": req.PvzID,
			"r.status": receipt.StatusInprogress,
		}).
		OrderBy("p.added_at DESC").
		Limit(1)

	sql, args, err := subQuery.ToSql()
	if err != nil {
		return 0, postgres.HandleError(err)
	}

	var productID uuid.UUID
	err = s.db.QueryRow(ctx, sql, args...).Scan(&productID)
	if err != nil {
		return 0, postgres.HandleError(err)
	}

	deleteBuilder := s.stmpBuilder().
		Delete(productsTable).
		Where(sq.Eq{"id": productID})

	sql, args, err = deleteBuilder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "can't build delete query")
	}

	result, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, postgres.HandleError(err)
	}

	return result.RowsAffected(), nil
}

func (s *storageImpl) GetLastReceipt(ctx context.Context, req receipt.GetLastReceiptQuery) (*receipt.Receipt, error) {
	queryBuilder := s.stmpBuilder().
		Select(receiptRowFields).
		From(receptionsTable).
		Where(sq.Eq{
			"pvz_id": req.PvzID,
			"status": receipt.StatusInprogress,
		}).
		OrderBy("reception_date DESC").
		Limit(1)

	if req.ForUpdate {
		queryBuilder = queryBuilder.Suffix("FOR UPDATE")
	}

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "can't build query")
	}
	var rr receiptRow
	err = s.db.QueryRow(ctx, sql, args...).Scan(&rr.ID, &rr.PvzID, &rr.ReceptionDate, &rr.Status)
	if err != nil {
		return nil, postgres.HandleError(err)
	}


	model := rr.ToModelReceipt()
	return &model, nil
}
