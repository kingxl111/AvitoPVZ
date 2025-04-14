package logging

import (
	"context"
	"log/slog"
)

// ExtractAttrsFunc - extract attrs from context.
type ExtractAttrsFunc func(ctx context.Context) []slog.Attr

// ExtractAttrsFromContext extract log attrs from context with extractor funcs.
func ExtractAttrsFromContext(ctx context.Context, extractors []ExtractAttrsFunc) []slog.Attr {
	return extractAttrsFromContext(ctx, extractors)
}

func extractAttrsFromContext(ctx context.Context, extractors []ExtractAttrsFunc) []slog.Attr {
	extracted := make([]slog.Attr, 0, len(extractors))
	for idx := range extractors {
		extracted = append(extracted, extractors[idx](ctx)...)
	}

	return extracted
}
