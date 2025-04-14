package user

import (
	"github.com/pkg/errors"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserAlreadyExist = errors.New("user already exists")
	ErrWrongPassword    = errors.New("wrong password")
)
