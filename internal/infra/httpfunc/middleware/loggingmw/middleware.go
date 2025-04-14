package loggingmw

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"AvitoPVZ/internal/infra/httpfunc/middleware"

	"AvitoPVZ/pkg/logging"
)

const (
	httpRequestMessage   = "http request"
	methodField          = "method"
	uriField             = "uri"
	remoteAddr           = "remote_addr"
	statusCodeField      = "status_code"
	sizeField            = "size"
	processTimeField     = "process_time"
	requestBodyField     = "request_body"
	requestSizeField     = "request_size"
	bodyField            = "body"
	headersField         = "headers"
	responseHeadersField = "response_headers"
)

type Options struct {
	sensitiveSkipper []middleware.ReplaceFunc
	extractors       []logging.ExtractAttrsFunc
	skippers         []middleware.SkipMiddlewareFunc
	logBody          bool
}

type Option func(o *Options)

// WithNoLogBody - don't log body.
func WithNoLogBody() Option {
	return func(o *Options) {
		o.logBody = false
	}
}

// WithSkipper - middleware skip logging funcs.
func WithSkipper(f ...middleware.SkipMiddlewareFunc) Option {
	return func(o *Options) {
		o.skippers = append(o.skippers, f...)
	}
}

// WithExtractors - logging context extractors.
func WithExtractors(f ...logging.ExtractAttrsFunc) Option {
	return func(o *Options) {
		o.extractors = append(o.extractors, f...)
	}
}

func WithSensitiveHeaders(f ...middleware.ReplaceFunc) Option {
	return func(o *Options) {
		o.sensitiveSkipper = append(o.sensitiveSkipper, f...)
	}
}

// LoggingMiddleware - logging middleware.
func LoggingMiddleware(ctx context.Context, lg *slog.Logger, opts ...Option) func(next http.Handler) http.Handler {
	op := Options{
		logBody: true,
	}

	for _, o := range opts {
		o(&op)
	}

	maskHeaders := middleware.MaskHeadersFunc(op.sensitiveSkipper)
	skipper := middleware.SkipperFunc(op.skippers)

	lgLevel := logging.LogLevel(ctx, lg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				childCtx := r.Context()

				if skipper(childCtx, r) {
					next.ServeHTTP(w, r)
					return
				}

				requestSize := r.ContentLength

				ctxFields := logging.ExtractAttrsFromContext(childCtx, op.extractors)
				fields := append(
					[]slog.Attr{
						slog.String(methodField, r.Method),
						slog.String(uriField, r.RequestURI),
						slog.String(remoteAddr, r.RemoteAddr),
						slog.Int64(requestSizeField, requestSize),
						slog.Any(headersField, maskHeaders(r.Header)),
					}, ctxFields...,
				)

				if op.logBody {
					bb, err := cloneBody(r)
					switch {
					case err != nil:
						lg.Error("cloneBody", slog.Any("err", err))
					default:
						fields = append(fields, slog.String(requestBodyField, bb.String()))
					}
				}

				rw := middleware.NewWrapResponseWriter(w)

				next.ServeHTTP(rw, r)

				responseSize := rw.Size()

				fields = append(
					fields, []slog.Attr{
						slog.Int(statusCodeField, rw.Status()),
						slog.Int(sizeField, responseSize),
						slog.Float64(processTimeField, time.Since(start).Seconds()),
						slog.Any(responseHeadersField, maskHeaders(rw.Header())),
					}...,
				)

				fields = append(fields, ctxFields...)
				if op.logBody {
					fields = append(fields, slog.String(bodyField, rw.String()))
				}

				lg.LogAttrs(childCtx, lgLevel, httpRequestMessage, fields...)
			},
		)
	}
}

func cloneBody(r *http.Request) (*bytes.Buffer, error) {
	var rb bytes.Buffer

	if r.Body == nil {
		return &rb, nil
	}

	if _, err := rb.ReadFrom(r.Body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	r.Body = io.NopCloser(bytes.NewReader(rb.Bytes()))

	return &rb, nil
}
