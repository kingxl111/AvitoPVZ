package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPingTimeout     = time.Second
	defaultMaxConnIdleTime = 5 * time.Minute
	defaultConnTimeout     = 30 // in seconds

	defaultMaxConns = 20 // nolint:unused
	defaultMinConns = 2  // nolint:unused
)

type config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	ConnTimeout     int
	MinConns        int
	MaxConns        int
	PingTimeout     time.Duration
	MaxConnIdleTime time.Duration
}

type Option func(*config)

func WithHost(host string) Option {
	return func(c *config) {
		c.Host = host
	}
}

func WithPort(port string) Option {
	return func(c *config) {
		c.Port = port
	}
}

func WithUser(user string) Option {
	return func(c *config) {
		c.User = user
	}
}

func WithPassword(password string) Option {
	return func(c *config) {
		c.Password = password
	}
}

func WithDBName(dbname string) Option {
	return func(c *config) {
		c.DBName = dbname
	}
}

func WithSSLMode(sslmode string) Option {
	return func(c *config) {
		c.SSLMode = sslmode
	}
}

func WithMaxConns(maxConns int) Option {
	return func(c *config) {
		c.MaxConns = maxConns
	}
}

func WithMinConns(minConns int) Option {
	return func(c *config) {
		c.MinConns = minConns
	}
}

func WithMaxConnIdleTime(maxConnIdleTime time.Duration) Option {
	return func(c *config) {
		c.MaxConnIdleTime = maxConnIdleTime
	}
}

func newConfig(opts ...Option) *config {
	cfg := &config{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "postgres",
		SSLMode:         "disable",
		ConnTimeout:     defaultConnTimeout,
		MinConns:        2,
		MaxConns:        10,
		PingTimeout:     defaultPingTimeout,
		MaxConnIdleTime: defaultMaxConnIdleTime,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type DB struct {
	*pgxpool.Pool
}

func New(ctx context.Context, opts ...Option) (*DB, error) {
	cfg := newConfig(opts...)
	// From pgx parse config
	//
	//   # Example DSN
	//   user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca pool_max_conns=10
	//
	//   # Example URL
	//   postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca&pool_max_conns=10
	pgxConfig, err := pgxpool.ParseConfig(pgxDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	// set BeforeAcquire helper func for pinging
	pgxConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		pCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
		defer cancel()

		return conn.Ping(pCtx) == nil
	}

	//fmt.Println("Creating connection beginning")
	//pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	//fmt.Printf("postgres://%s:%s@%s:%s/%s?sslmode=%s\n", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func pgxDSN(cfg *config) string {
	buf := strings.Builder{}

	fmt.Fprintf(&buf, "dbname=%s ", cfg.DBName)
	fmt.Fprintf(&buf, "user=%s ", cfg.User)
	fmt.Fprintf(&buf, "password=%s ", cfg.Password)
	fmt.Fprintf(&buf, "host=%s ", cfg.Host)
	fmt.Fprintf(&buf, "port=%s ", cfg.Port)
	fmt.Fprintf(&buf, "sslmode=%s ", cfg.SSLMode)
	fmt.Fprintf(&buf, "connect_timeout=%d ", cfg.ConnTimeout)
	fmt.Fprintf(&buf, "pool_min_conns=%d ", cfg.MinConns)
	fmt.Fprintf(&buf, "pool_max_conns=%d ", cfg.MaxConns)

	return buf.String()
}
