package main

import (
	"context"
	"log/slog"

	"AvitoPVZ/pkg/logging"
)

type requestContextKey string

const contextKeyRequestID = requestContextKey("request_id")

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}

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

func main() {
	extractRequestID := func(ctx context.Context) []slog.Attr {
		if id := RequestIDFromContext(ctx); id != "" {
			return []slog.Attr{slog.String("request_id", id)}
		}

		return nil
	}

	lg, err := logging.New(
		logging.WithLevel(slog.LevelDebug.String()),
		logging.WithJSONHandler(), logging.WithDebugInfo(),
		logging.WithHandleMiddleware(
			func(h slog.Handler) slog.Handler {
				return logging.NewCtxHandler(h, extractRequestID)
			},
		),
	)
	if err != nil {
		panic(err)
	}

	ctx := withRequestID(context.Background(), "123")

	lg1 := lg.With("op", "123")

	lg1.InfoContext(ctx, "hello world")
}
