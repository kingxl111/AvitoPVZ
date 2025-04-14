package requestidmw

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type requestContextKey string

const (
	contextKeyRequestID = requestContextKey("request_id")
	xRequestIDHeader    = "X-Request-ID"
)

// PopulateRequestID creates a new context with request id.
// If the request id is not passed in the http header, a new request id is generated.
func PopulateRequestID(lg *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				requestID := r.Header.Get(xRequestIDHeader)
				if requestID == "" {
					u, err := uuid.NewRandom()
					if err != nil {
						lg.Error("uuid.NewRandom", slog.Any("err", err))
						next.ServeHTTP(w, r)

						return
					}
					requestID = u.String()
				}

				ctx1 := withRequestID(ctx, requestID)
				r1 := r.Clone(ctx1)
				next.ServeHTTP(w, r1)
			},
		)
	}
}

// RequestIDFromContext return request id or empty string.
func RequestIDFromContext(ctx context.Context) string {
	v := ctx.Value(contextKeyRequestID)
	if v == nil {
		return ""
	}

	t, ok := v.(string)
	if !ok {
		return ""
	}
	return t
}

// withRequestID create new context with id.
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}
