package story

import (
	"AvitoPVZ/internal/storage"
	"AvitoPVZ/internal/user"
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (s *Story) Login(ctx context.Context, req user.LoginRequest) (string, error) {
	u, err := s.storage.GetUser(ctx, user.GetUserQuery{
		Email: req.Email,
	})
	if err != nil {
		return "", errors.WithMessage(err, "CreateUser")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "bcrypt.GenerateFromPassword")
	}

	if u.HashedPwd != string(hashedPassword) {
		return "", user.ErrWrongPassword
	}

	token, err := user.GenerateToken(u.ID, u.Role)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", user.ErrUserNotFound
		}
		return "", errors.WithMessage(err, "GenerateToken")
	}

	return token, nil
}
