package story

import (
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
		return "", errors.WithMessage(err, "GetUser")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPwd), []byte(req.Password))
	if err != nil {
		return "", user.ErrWrongPassword
	}

	token, err := user.GenerateToken(u.ID, u.Role)
	if err != nil {
		return "", errors.WithMessage(err, "GenerateToken")
	}

	return token, nil
}
