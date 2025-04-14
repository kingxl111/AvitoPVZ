package environment

import (
	"context"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type ServerOptions struct {
	logger        *slog.Logger
	panicHandler  func(w http.ResponseWriter, r *http.Request, p any)
	middlewares   []func(http.Handler) http.Handler
	serverOptions []func(*http.Server)
}

func (o *ServerOptions) WithLogger(logger *slog.Logger) {
	o.logger = logger
}

func (o *ServerOptions) WithPanicHandler(h func(w http.ResponseWriter, r *http.Request, p any)) {
	o.panicHandler = h
}

func (o *ServerOptions) WithServerOptions(v ...func(*http.Server)) {
	o.serverOptions = append(o.serverOptions, v...)
}

func (o *ServerOptions) WithMiddlewares(v ...func(http.Handler) http.Handler) {
	o.middlewares = append(o.middlewares, v...)
}

func (o *ServerOptions) NewServer(handler http.Handler, addr string) *http.Server {
	if o.logger == nil {
		o.logger = slog.Default()
	}

	if o.panicHandler == nil {
		o.panicHandler = func(w http.ResponseWriter, r *http.Request, p any) {
			o.logger.Error("recovered from panic",
				"panic", p,
				"stack", debug.Stack(),
				"method", r.Method,
				"path", r.URL.Path,
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

	wrappedHandler := handler
	for _, mw := range o.middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	wrappedHandler = o.loggingMiddleware(wrappedHandler)
	wrappedHandler = o.recoveryMiddleware(wrappedHandler)

	srv := &http.Server{
		Handler: wrappedHandler,
		Addr:    addr,
	}

	for _, opt := range o.serverOptions {
		opt(srv)
	}

	return srv
}

func ListenAndServeContext(ctx context.Context, srv *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (o *ServerOptions) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		o.logger.Info("request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		lw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lw, r)

		o.logger.Info("request completed",
			"status", lw.status,
			"duration", time.Since(start),
			"size", lw.size,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (l *loggingResponseWriter) WriteHeader(status int) {
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func (o *ServerOptions) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				o.panicHandler(w, r, p)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
