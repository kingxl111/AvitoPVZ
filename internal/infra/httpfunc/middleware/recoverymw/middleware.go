package recoverymw

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"AvitoPVZ/pkg/logging"
)

// Recovery - recovery from panic middleware.
func Recovery(lg *slog.Logger, extractors ...logging.ExtractAttrsFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if p := recover(); p != nil {
						lgLevel := logging.LogLevel(r.Context(), lg)
						lg.LogAttrs(
							r.Context(), lgLevel, "http handler panic", append(
								[]slog.Attr{slog.String("stack", string(debug.Stack()))},
								logging.ExtractAttrsFromContext(r.Context(), extractors)...,
							)...,
						)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}()

				next.ServeHTTP(w, r)
				return
			},
		)
	}
}
