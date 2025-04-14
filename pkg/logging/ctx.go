package logging

import (
	"context"
	"log/slog"
)

func NewCtxHandler(next slog.Handler, extractors ...ExtractAttrsFunc) slog.Handler {
	return &ctxHandler{next: next, extractors: extractors}
}

type ctxHandler struct {
	next       slog.Handler
	extractors []ExtractAttrsFunc
}

func (h *ctxHandler) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

// Handle - handle log record and add attrs from context.
func (h *ctxHandler) Handle(ctx context.Context, rec slog.Record) error {
	rec.AddAttrs(extractAttrsFromContext(ctx, h.extractors)...)
	return h.next.Handle(ctx, rec)
}

func (h *ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{
		next:       h.next.WithAttrs(attrs),
		extractors: append(h.extractors[:0:0], h.extractors...),
	}
}

func (h *ctxHandler) WithGroup(name string) slog.Handler {
	return &ctxHandler{
		next:       h.next.WithGroup(name),
		extractors: append(h.extractors[:0:0], h.extractors...),
	}
}
