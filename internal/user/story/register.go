package story

import (
	"AvitoPVZ/internal/storage"
	"AvitoPVZ/internal/user"
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (s *Story) Register(ctx context.Context, req user.RegisterUserRequest) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "bcrypt.GenerateFromPassword")
	}

	u, err := s.storage.InsertUser(ctx, user.InsertUserQuery{
		Email:     req.Email,
		HashedPwd: string(hashedPassword),
		Role:      req.Role,
	})
	if err != nil {
		if errors.Is(err, storage.ErrEntityAlreadyExist) {
			return nil, user.ErrUserAlreadyExist
		}
		return nil, errors.WithMessage(err, "CreateUser")
	}

	return u, nil
}
