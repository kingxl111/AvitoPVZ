package xpgx

import (
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

var (
	// ErrNotFound ошибка при отсутствии сущности.
	ErrNotFound = errors.New("entity not found")
	// ErrEntityAlreadyExist ошибка при попытке создать сущность, которая уже существует.
	ErrEntityAlreadyExist = errors.New("entity already exists")
	// ErrLockedEntityAccess ошибка при попытке получить доступ к заблокированной сущности.
	ErrLockedEntityAccess = errors.New("an attempt to acquire an access to the locked entity")
)

// HandleError кастомная обработка ошибок postgres.
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
		return ErrNotFound
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
		return ErrEntityAlreadyExist
	case PgErrConcurrentLockAcquisition:
		return ErrLockedEntityAccess
	default:
		return err
	}
}
