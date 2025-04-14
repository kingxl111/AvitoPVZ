package environment

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/infra/postgres"
	"context"

	"github.com/pkg/errors"
)

type Clients struct {
	DB *postgres.DB
}

func newClients(ctx context.Context, cfg config.Config) (*Clients, error) {
	var clts Clients

	db, err := postgres.New(
		ctx,
		postgres.WithDBName(cfg.DB.DBName),
		postgres.WithHost(cfg.DB.Host),
		postgres.WithPort(cfg.DB.Port),
		postgres.WithUser(cfg.DB.User),
		postgres.WithPassword(cfg.DB.Password),
	)
	if err != nil {
		return nil, errors.Wrap(err, "postgres.New")
	}

	if err = db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "db.Ping")
	}

	clts.DB = db

	return &clts, nil
}
