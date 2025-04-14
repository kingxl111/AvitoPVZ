package pvz

import (
	"github.com/pkg/errors"
)

var (
	ErrPVZAlreadyExist = errors.New("pvz already exists")
	ErrPVZNotFound     = errors.New("pvz not found")
)
