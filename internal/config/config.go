package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"time"
)

func Load(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}

const JwtSecret = "JWT_SECRET"

type Config struct {
	Env              string                  `env:"ENV,default=local"`
	Logger           LoggerConfig            `env:",prefix=LOGGER_"`
	Observability    ObservabilityHTTPConfig `env:",prefix=OBSERVABILITY_"`
	APIServer        APIServerHTTPConfig     `env:",prefix=API_SERVER_"`
	ShutdownDuration time.Duration           `env:"SHUTDOWN_DURATION,default=30s"`
	DB               PostgresConfig          `env:",prefix=PG_"`
	Metrics          struct {
		Collector struct {
			Timeout time.Duration `env:"COLLECTOR_TIMEOUT,default=10s"`
		} `env:",prefix=COLLECTOR_"`
	} `env:"METRICS"`
}

type APIServerHTTPConfig struct {
	Host              string        `env:"HOST,default=127.0.0.1"`
	Port              uint16        `env:"PORT,default=8080"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT,default=30s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT,default=30s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT,default=30s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT,default=30s"`
	MaxBodyBytes      int64         `env:"MAX_BODY_BYTES,default=1048576"`
}

func (a APIServerHTTPConfig) ADDR() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

type HTTPClientConfig struct {
	Scheme        string        `env:"SCHEME,default=http"`
	Host          string        `env:"HOST,default=127.0.0.1"`
	Port          uint16        `env:"PORT,default=9000"`
	Timeout       time.Duration `env:"TIMEOUT,default=30s"`
	MaxRetries    int           `env:"MAX_RETRIES,default=3"`
	RetryInterval time.Duration `env:"RETRY_INTERVAL,default=2s"`
	RateLimit     struct {
		Burst int     `env:"BURST,default=0"`
		RPS   float64 `env:"RPS,default=20.0"`
	} `env:",prefix=RATE_LIMIT_"`
}

func (c HTTPClientConfig) ADDR() string {
	return fmt.Sprintf("%s://%s:%d", c.Scheme, c.Host, c.Port)
}

type ObservabilityHTTPConfig struct {
	Host         string        `env:"HOST,default=127.0.0.1"`
	Port         uint16        `env:"PORT,default=8383"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT,default=30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT,default=30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT,default=1m"`
}

func (a ObservabilityHTTPConfig) ADDR() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

type PostgresConfig struct {
	Host     string `env:"HOST,default=localhost"`
	Port     string `env:"PORT,default=5432"`
	DBName   string `env:"DB_NAME,default=dm_live"`
	User     string `env:"USER,default=postgres"`
	Password string `env:"PASSWORD,default=postgres"`
}
