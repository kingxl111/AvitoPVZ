package routes

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/infra/httpfunc/middleware"
	"AvitoPVZ/internal/infra/httpfunc/middleware/authmw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/loggingmw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/recoverymw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/requestidmw"
	"AvitoPVZ/internal/infra/httpfunc/middleware/timeoutmw"
	"AvitoPVZ/internal/ingress/gates/apihandler"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Router(
	ctx context.Context,
	cfg config.Config,
	lg *slog.Logger,
	handler *apihandler.Handler,
) http.Handler {
	router := chi.NewRouter()

	extractRequestID := func(ctx context.Context) []slog.Attr {
		if id := requestidmw.RequestIDFromContext(ctx); id != "" {
			return []slog.Attr{slog.String("request_id", id)}
		}
		return nil
	}

	router.Use(
		requestidmw.PopulateRequestID(lg),
		timeoutmw.TimeoutMiddleware(timeoutmw.WithTimeout(30*time.Second)),
		recoverymw.Recovery(lg, extractRequestID),
		loggingmw.LoggingMiddleware(
			ctx, lg,
			loggingmw.WithSkipper(middleware.PathSkipper("/pvzops/health")),
			loggingmw.WithExtractors(extractRequestID),
			loggingmw.WithSensitiveHeaders(middleware.BasicAuthReplacer),
		),
	)

	router.Group(func(r chi.Router) {
		r.Post("/dummyLogin", handler.PostDummyLogin)
		r.Post("/register", handler.PostRegister)
		r.Post("/login", handler.PostLogin)
	})

	router.Group(func(r chi.Router) {
		r.Use(authmw.AuthMiddleware)

		r.Route("/pvz", func(pvzRouter chi.Router) {
			pvzRouter.Post("/", func(writer http.ResponseWriter, request *http.Request) {})
			pvzRouter.Get("/", func(writer http.ResponseWriter, request *http.Request) {})
			pvzRouter.Post("/{pvzId}/close_last_reception", func(writer http.ResponseWriter, request *http.Request) {})
			pvzRouter.Post("/{pvzId}/delete_last_product", func(writer http.ResponseWriter, request *http.Request) {})
		})

		r.Post("/receptions", func(writer http.ResponseWriter, request *http.Request) {})
		r.Post("/products", func(writer http.ResponseWriter, request *http.Request) {})
	})

	router.HandleFunc("/pvzops/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return router
}
