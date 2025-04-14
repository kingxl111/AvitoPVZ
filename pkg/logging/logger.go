package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
	With(args ...any) *slog.Logger
	WithGroup(name string) *slog.Logger
	ErrorContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
}

type MiddlewareFunc func(h slog.Handler) slog.Handler

const (
	defaultHandleTyp = handleTypText
	defaultLevel     = "DEBUG"
	defaultDebuginfo = false
)

type handleTyp string

func (h handleTyp) isText() bool { // nolint:unused
	return h == handleTypText
}

func (h handleTyp) isJSON() bool {
	return h == handleTypJSON
}

const (
	handleTypText handleTyp = "text"
	handleTypJSON handleTyp = "json"
)

type Options struct {
	debugInfo  bool
	level      string
	handler    handleTyp
	source     bool
	middleware []MiddlewareFunc
}

type Option func(o *Options)

func WithDebugInfo() Option {
	return func(o *Options) {
		o.debugInfo = true
	}
}

func WithLevel(l string) Option {
	return func(o *Options) {
		o.level = l
	}
}

func WithTextHandler() Option {
	return func(o *Options) {
		o.handler = handleTypText
	}
}

func WithJSONHandler() Option {
	return func(o *Options) {
		o.handler = handleTypJSON
	}
}

func WithAddSource() Option {
	return func(o *Options) {
		o.source = true
	}
}

func WithHandleMiddleware(f ...MiddlewareFunc) Option {
	return func(o *Options) {
		o.middleware = append(o.middleware, f...)
	}
}

func New(opts ...Option) (*slog.Logger, error) {
	options := Options{
		debugInfo: defaultDebuginfo,
		level:     defaultLevel,
		handler:   defaultHandleTyp,
	}

	for _, o := range opts {
		o(&options)
	}

	l := &slog.LevelVar{}
	if err := l.UnmarshalText([]byte(options.level)); err != nil {
		return nil, fmt.Errorf("slog UnmarshalText: %w", err)
	}

	handleOpts := &slog.HandlerOptions{
		AddSource: options.source,
		Level:     l,
	}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, handleOpts)
	if options.handler.isJSON() {
		h = slog.NewJSONHandler(os.Stdout, handleOpts)
	}

	for _, m := range options.middleware {
		h = m(h)
	}

	return slog.New(h), nil
}

func Development() *slog.Logger {
	l, _ := New(WithLevel("DEBUG"), WithTextHandler(), WithDebugInfo())

	return l
}

func Production() *slog.Logger {
	l, _ := New(WithLevel("INFO"), WithJSONHandler(), WithDebugInfo())

	return l
}

func LogLevel(ctx context.Context, lg *slog.Logger) slog.Level {
	var level slog.Level

	switch {
	case lg.Enabled(ctx, slog.LevelDebug):
		level = slog.LevelDebug
	case lg.Enabled(ctx, slog.LevelInfo):
		level = slog.LevelInfo
	case lg.Enabled(ctx, slog.LevelWarn):
		level = slog.LevelWarn
	case lg.Enabled(ctx, slog.LevelError):
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	return level
}
