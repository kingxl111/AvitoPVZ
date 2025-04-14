package product

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeElectronics Type = "электроника"
	TypeClothing    Type = "одежда"
	TypeShoes       Type = "обувь"
)

type Product struct {
	ID       uuid.UUID
	PvzID    uuid.UUID
	Type     Type
	DateTime time.Time
}

type AddProductRequest struct {
	Type  Type
	PvzID uuid.UUID
}

type InsertProductQuery struct {
	Type  Type
	PvzID uuid.UUID
}

type DeleteLastProductRequest struct {
	PvzID uuid.UUID
}

type DeleteLastProductQuery struct {
	PvzID uuid.UUID
}
