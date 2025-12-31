package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/yourorg/anonymous-support/internal/pkg/metrics"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode)).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		})
	}
}
