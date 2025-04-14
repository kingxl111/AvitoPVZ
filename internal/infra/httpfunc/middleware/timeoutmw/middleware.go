package timeoutmw

import (
	"context"
	"errors"
	"net/http"
	"time"

	"AvitoPVZ/internal/infra/httpfunc/middleware"
)

var defaultTimeout = 30 * time.Second

const ClientRequestCloseStatusCode = 499

type Option func(options *Options)

type Options struct {
	timeout    time.Duration
	statusCode int
	skippers   []middleware.SkipMiddlewareFunc
}

func WithTimeout(t time.Duration) Option {
	return func(options *Options) {
		options.timeout = t
	}
}

func WithStatusCode(statusCode int) Option {
	return func(options *Options) {
		options.statusCode = statusCode
	}
}

// WithSkipper - middleware skip logging funcs.
func WithSkipper(f ...middleware.SkipMiddlewareFunc) Option {
	return func(o *Options) {
		o.skippers = append(o.skippers, f...)
	}
}

// TimeoutMiddleware - middleware for cancel context.
func TimeoutMiddleware(opts ...Option) func(next http.Handler) http.Handler {
	opt := Options{
		timeout:    defaultTimeout,
		statusCode: ClientRequestCloseStatusCode,
	}

	for _, o := range opts {
		o(&opt)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				ctx, cancel := context.WithTimeout(ctx, opt.timeout)
				defer cancel()

				doneCh := make(chan struct{})

				rw := middleware.NewWrapResponseWriter(w)

				go func() {
					defer close(doneCh)

					r = r.WithContext(ctx)
					next.ServeHTTP(rw, r)
				}()

				select {
				case <-ctx.Done():
					switch {
					case errors.Is(ctx.Err(), context.DeadlineExceeded):
						rw.WriteHeader(http.StatusGatewayTimeout)
					case errors.Is(ctx.Err(), context.Canceled):
						rw.WriteHeader(opt.statusCode)
					default:
					}
				case <-doneCh:
				}
			},
		)
	}
}
