package pgstore

import (
	"AvitoPVZ/internal/infra/postgres"
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"AvitoPVZ/internal/user"
)

const usersTable = "users"

var userRowFields = fields(userRow{})

type userRow struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	HashedPassword string    `db:"hashed_password"`
	Role           string    `db:"role"`
	CreatedAt      time.Time `db:"created_at"`
}

func (ur userRow) ToModel() *user.User {
	return &user.User{
		ID:        ur.ID,
		Email:     ur.Email,
		HashedPwd: ur.HashedPassword,
		Role:      user.Role(ur.Role),
	}
}

func (s *storageImpl) InsertUser(ctx context.Context, req user.InsertUserQuery) (*user.User, error) {
	params := map[string]interface{}{
		"email":           req.Email,
		"hashed_password": req.HashedPwd,
		"role":            req.Role,
		"created_at":      s.now(),
	}

	q, args, err := s.stmpBuilder().
		Insert(usersTable).
		SetMap(params).
		Suffix("RETURNING " + userRowFields).
		ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	ur, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	return ur.ToModel(), nil
}

func (s *storageImpl) GetUser(ctx context.Context, req user.GetUserQuery) (*user.User, error) {
	q, args, err := s.stmpBuilder().
		Select(userRowFields).
		From(usersTable).
		Where(sq.Eq{"email": req.Email}).
		ToSql()
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, postgres.HandleError(err)
	}
	defer rows.Close()

	ur, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return nil, postgres.HandleError(err)
	}

	return ur.ToModel(), nil
}
