package main

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/environment"
	"AvitoPVZ/internal/infra/postgres"
	"AvitoPVZ/internal/ingress/gates/apihandler"
	prodStory "AvitoPVZ/internal/product/story"
	pvzStory "AvitoPVZ/internal/pvz/story"
	receiptStory "AvitoPVZ/internal/receipt/story"
	"AvitoPVZ/internal/storage/pgstore"
	userStory "AvitoPVZ/internal/user/story"
	"AvitoPVZ/pkg/api/oapigen/pvzops"
	"AvitoPVZ/pkg/xpgx"
	"AvitoPVZ/pkg/xpgx/transaction"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	defaultLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(defaultLogger)

	if err := runMain(ctx); err != nil {
		defaultLogger.Error("run main", slog.Any("err", err))
		return
	}
}

func runMain(ctx context.Context) error {
	flag.Parse()

	if err := config.Load(configPath); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	pgConfig, err := config.NewPGConfig()
	if err != nil {
		return fmt.Errorf("pg config: %w", err)
	}

	db, err := postgres.New(
		ctx,
		postgres.WithUser(pgConfig.Username),
		postgres.WithPassword(pgConfig.Password),
		postgres.WithHost(pgConfig.Host),
		postgres.WithPort(pgConfig.Port),
		postgres.WithDBName(pgConfig.DBName),
		postgres.WithSSLMode(pgConfig.SSLMode),
	)
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	defer db.Close()

	loggerConfig, err := config.NewLoggerConfig()
	if err != nil {
		return fmt.Errorf("failed to get logger config: %v", err)
	}

	handleOpts := &slog.HandlerOptions{
		Level: loggerConfig.Level(),
	}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, handleOpts)
	logger := slog.New(h)

	repo := pgstore.New(db)
	txManager := transaction.NewManager((*xpgx.DB)(db))
	now := func() time.Time { return time.Now().UTC() }
	userSt := userStory.New(repo)
	pvzSt := pvzStory.New(repo, txManager)
	receiptSt := receiptStory.New(repo, now, txManager)
	productSt := prodStory.New(repo, txManager)

	httpServerConfig, err := config.NewHTTPConfig()
	if err != nil {
		return fmt.Errorf("http server config error: %w", err)
	}

	var opts environment.ServerOptions
	opts.WithLogger(logger)
	handler := apihandler.New(logger, userSt, pvzSt, productSt, receiptSt)
	mux := http.NewServeMux()
	apiHandler := pvzops.HandlerFromMux(handler, mux)
	httpServer := opts.NewServer(apiHandler, httpServerConfig.Address())

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		logger.Info("starting http server on " + httpServerConfig.Address() + "...")
		if err := environment.ListenAndServeContext(ctx, httpServer); err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error during server shutdown", slog.Any("err", err))
			return err
		}

		logger.Info("server stopped gracefully")
		return nil
	})

	if err := eg.Wait(); err != nil {
		logger.Error("server terminated with error", slog.Any("err", err))
		return fmt.Errorf("server terminated with error: %w", err)
	}

	logger.Info("server exited cleanly")
	return nil
}
