package story

import (
	"AvitoPVZ/internal/user"
	"context"
)

//go:generate mockgen -source=contracts.go -destination=mocks.go -package=story
type (
	stg interface {
		GetUser(ctx context.Context, req user.GetUserQuery) (*user.User, error)
		InsertUser(ctx context.Context, req user.InsertUserQuery) (*user.User, error)
	}
)
