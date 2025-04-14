package environment

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/pkg/logging"
	"context"
	"fmt"
	"log/slog"

	"github.com/sethvargo/go-envconfig"
)

type Env struct {
	Config  *config.Config
	Logger  *slog.Logger
	Clients *Clients
}

func Setup(ctx context.Context) (*Env, error) {
	var cfg config.Config
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return nil, fmt.Errorf("env processing: %w", err)
	}

	var e Env

	logger, err := initLogger(cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("initLogger: %w", err)
	}

	clts, err := newClients(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("newClients: %w", err)
	}

	e.Config = &cfg
	e.Logger = logger
	e.Clients = clts

	return &e, nil
}

func initLogger(cfg config.LoggerConfig) (*slog.Logger, error) {
	lgOpts := make([]logging.Option, 0)
	lgOpts = append(lgOpts, logging.WithLevel((string)(cfg.Level())))

	logger, err := logging.New(append(lgOpts, logging.WithDebugInfo())...)
	if err != nil {
		return nil, fmt.Errorf("logging.New: %w", err)
	}

	return logger, nil
}
