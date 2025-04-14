package config

import (
	"os"
)

const (
	dsnEnvName = "PG_DSN"
	user       = "DATABASE_USER"
	password   = "DATABASE_PASSWORD"
	name       = "DATABASE_NAME"
	host       = "DATABASE_HOST"
	port       = "DATABASE_PORT"
	sslmode    = "disable"
)

type PGConfig struct {
	dsn      string
	DBName   string
	Username string
	Password string
	Host     string
	Port     string
	SSLMode  string
}

func NewPGConfig() (*PGConfig, error) {
	var cfg PGConfig
	cfg.DBName = os.Getenv(name)
	cfg.Username = os.Getenv(user)
	cfg.Password = os.Getenv(password)
	cfg.Host = os.Getenv(host)
	cfg.Port = os.Getenv(port)
	cfg.SSLMode = sslmode
	cfg.dsn = os.Getenv(dsnEnvName)

	return &cfg, nil
}
