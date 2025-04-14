package story

import (
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/storage"
	"context"

	"github.com/pkg/errors"
)

func (s *Story) Create(ctx context.Context, req pvz.CreatePVZRequest) (*pvz.PVZ, error) {
	insertedPVZ, err := s.storage.InsertPVZ(ctx, pvz.InsertPVZQuery{
		ID:               req.ID,
		RegistrationDate: req.RegistrationDate,
		City:             req.City,
	})
	if err != nil {
		if errors.Is(err, storage.ErrEntityAlreadyExist) {
			return nil, pvz.ErrPVZAlreadyExist
		}
		return nil, errors.WithMessage(err, "storage.InsertPVZ")
	}

	return insertedPVZ, nil
}
