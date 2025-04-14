package story

import (
	"time"
)

type Story struct {
	storage   stg
	now       func() time.Time
	txManager manager
}

func New(storage stg, now func() time.Time, txManager manager) *Story {
	return &Story{
		storage:   storage,
		now:       now,
		txManager: txManager,
	}
}
