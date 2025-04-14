package pgstore

import (
	"AvitoPVZ/internal/infra/postgres"
	"reflect"
	"time"

	"github.com/Masterminds/squirrel"
)

type storageImpl struct {
	db  postgres.Executor
	now func() time.Time
}

func New(db postgres.Executor) *storageImpl { // nolint
	return &storageImpl{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *storageImpl) stmpBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func fields(data interface{}) string {
	var s string
	r := reflect.TypeOf(data)
	for i := 0; i < r.NumField(); i++ {
		tag := r.Field(i).Tag.Get("db")
		if tag != "" {
			s += tag + ","
		}
	}
	return s[:len(s)-1]
}
