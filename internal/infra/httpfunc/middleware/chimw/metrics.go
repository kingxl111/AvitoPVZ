package chimw

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"AvitoPVZ/internal/infra/httpfunc/middleware"

	"github.com/go-chi/chi/v5" // nolint
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// default chi prometheus chimw names.
const (
	requestsName = "requests_total"
	latencyName  = "requests_duration_seconds"
)

// default chi prometheus chimw metrics.
var (
	requestsTotalCounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: requestsName,
			Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
		}, []string{"code", "status", "method", "path"},
	)
	requestsLatencyHistogramVec = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: latencyName,
			Help: "How long it took to process the request, partitioned by status code, method and HTTP path.",
		}, []string{"code", "status", "method", "path"},
	)
)

// MetricsMiddlewareOpts - configuration options for custom chimw builder.
type MetricsMiddlewareOpts struct {
	GroupPathPattern            bool
	RequestsTotalCounterVec     *prometheus.CounterOpts
	RequestsLatencyHistogramVec *prometheus.HistogramOpts
}

// MetricsMiddlewareBuilder - build chimw with custom metrics.
func MetricsMiddlewareBuilder(opts MetricsMiddlewareOpts) func(next http.Handler) http.Handler {
	var (
		requestsTotal   *prometheus.CounterVec
		requestsLatency *prometheus.HistogramVec
	)

	requestsTotal = requestsTotalCounterVec
	requestsLatency = requestsLatencyHistogramVec

	if opts.RequestsTotalCounterVec != nil {
		requestsTotal = promauto.NewCounterVec(
			*opts.RequestsTotalCounterVec, []string{"code", "status", "method", "path"},
		)
	}

	if opts.RequestsLatencyHistogramVec != nil {
		requestsLatency = promauto.NewHistogramVec(
			*opts.RequestsLatencyHistogramVec, []string{"code", "status", "method", "path"},
		)
	}

	// resolve chimw type
	if opts.GroupPathPattern {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					start := time.Now()

					ww := middleware.NewWrapResponseWriter(w)
					next.ServeHTTP(ww, r)

					rctx := chi.RouteContext(r.Context())
					// replace routing patterns
					// Example  /updates/{id} instead of /updates/a995eace-4a41-11ed-a260-e2927d83fba9
					routePattern := strings.Join(rctx.RoutePatterns, "")
					routePattern = strings.Replace(routePattern, "/*/", "/", -1)

					code := strconv.Itoa(ww.Status())
					status := http.StatusText(ww.Status())
					method := r.Method
					path := routePattern

					requestsTotal.WithLabelValues(code, status, method, path).Inc()
					requestsLatency.
						WithLabelValues(code, status, method, path).
						Observe(time.Since(start).Seconds())
				},
			)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				ww := middleware.NewWrapResponseWriter(w)
				next.ServeHTTP(ww, r)

				code := strconv.Itoa(ww.Status())
				status := http.StatusText(ww.Status())
				method := r.Method
				path := r.URL.Path

				requestsTotal.WithLabelValues(code, status, method, path).Inc()
				requestsLatency.
					WithLabelValues(code, status, method, path).
					Observe(time.Since(start).Seconds())
			},
		)
	}
}
