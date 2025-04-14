package postgres

import (
	"AvitoPVZ/internal/storage"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	// PgErrCodeUniqueViolation код ошибки Postgres при нарушении уникальности.
	PgErrCodeUniqueViolation = "23505"
	// PgErrConcurrentLockAcquisition код ошибки Postgres при попытке получить доступ к заблокированной сущности.
	PgErrConcurrentLockAcquisition = "55P03"
)

// HandleError кастомная обработка ошибок postgres.
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
		return storage.ErrNotFound
	}

	var (
		code      string
		pgConnErr *pgconn.PgError
		pgErr     *pq.Error
	)
	if ok := errors.As(err, &pgConnErr); ok {
		code = pgConnErr.Code
	}
	if ok := errors.As(err, &pgErr); ok {
		code = string(pgErr.Code)
	}

	switch code {
	case PgErrCodeUniqueViolation:
		return storage.ErrEntityAlreadyExist
	case PgErrConcurrentLockAcquisition:
		return storage.ErrLockedEntityAccess
	default:
		return err
	}
}
