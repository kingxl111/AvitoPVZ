package storage

import (
	"errors"
)

var (
	// ErrNotFound ошибка при отсутствии сущности.
	ErrNotFound = errors.New("entity not found")
	// ErrEntityAlreadyExist ошибка при попытке создать сущность, которая уже существует.
	ErrEntityAlreadyExist = errors.New("entity already exists")
	// ErrLockedEntityAccess ошибка при попытке получить доступ к заблокированной сущности.
	ErrLockedEntityAccess = errors.New("an attempt to acquire an access to the locked entity")
)
