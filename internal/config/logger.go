package config

import (
	"fmt"
	"log/slog"
	"os"
)

var _ LoggerConfig = (*loggerConfig)(nil)

const (
	loggerLevel = "LOGGER_LEVEL"
)

type LoggerConfig interface {
	Level() slog.Level
}

type loggerConfig struct {
	level slog.Level
}

func (c *loggerConfig) Level() slog.Level {
	return c.level
}

func NewLoggerConfig() (LoggerConfig, error) {
	lvl := os.Getenv(loggerLevel)

	var l slog.Level

	if err := l.UnmarshalText([]byte(lvl)); err != nil {
		return nil, fmt.Errorf("slog UnmarshalText: %w", err)
	}

	return &loggerConfig{
		level: l,
	}, nil
}

