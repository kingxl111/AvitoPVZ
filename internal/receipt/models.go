package receipt

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusInprogress Status = "in_progress"
	StatusClose      Status = "close"
)

type Receipt struct {
	ID       uuid.UUID
	PvzID    uuid.UUID
	DateTime time.Time
	Status   Status
}

type CreateReceiptRequest struct {
	PvzID uuid.UUID
}

type CloseLastReceiptRequest struct {
	PvzID uuid.UUID
}

type InsertReceiptQuery struct {
	PvzID    uuid.UUID
	DateTime time.Time
	Status   Status
}

type UpdateLastReceiptQuery struct {
	PvzID  uuid.UUID
	Status Status
}

type ListReceptionsQuery struct {
	PvzIDs    []uuid.UUID
	StartDate *string
	EndDate   *string
}

type GetLastReceiptQuery struct {
	PvzID     uuid.UUID
	ForUpdate bool
}
