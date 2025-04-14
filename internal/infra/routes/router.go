package routes

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/infra/httpfunc/middleware"
	"AvitoPVZ/internal/infra/httpfunc/middleware/chimw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/loggingmw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/recoverymw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/requestidmw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/timeoutmw"
	"AvitoPVZ/pkg/api/oapigen/pvzops"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Router(
	ctx context.Context,
	_ config.Config,
	lg *slog.Logger,
	apiHandler pvzops.ServerInterface,
) http.Handler {
	router := chi.NewRouter()

	extractRequestID := func(ctx context.Context) []slog.Attr {
		if id := requestidmw.RequestIDFromContext(ctx); id != "" {
			return []slog.Attr{slog.String("request_id", id)}
		}
		return nil
	}

	const timeout = 30 * time.Second

	router.Use(
		requestidmw.PopulateRequestID(lg),
		timeoutmw.TimeoutMiddleware(timeoutmw.WithTimeout(timeout)),
		recoverymw.Recovery(lg, extractRequestID),
		loggingmw.LoggingMiddleware(
			ctx, lg,
			loggingmw.WithSkipper(middleware.PathSkipper("/pvzops/health")),
			loggingmw.WithExtractors(extractRequestID),
			loggingmw.WithSensitiveHeaders(middleware.BasicAuthReplacer),
		),
		chimw.MetricsMiddlewareBuilder(
			chimw.MetricsMiddlewareOpts{
				GroupPathPattern: true,
			},
		),
	)
	router.Mount(
		"/pvzops/v1", pvzops.HandlerWithOptions(
			apiHandler, pvzops.StdHTTPServerOptions{
				ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					lg.Error("error handler func", "error", err)
					w.WriteHeader(http.StatusBadRequest)
				},
			},
		),
	)
	router.HandleFunc(
		"/pvzops/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	return router
}
