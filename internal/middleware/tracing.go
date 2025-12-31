package middleware

import (
	"net/http"

	"github.com/yourorg/anonymous-support/internal/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adds OpenTelemetry tracing to HTTP requests
func TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start a new span for this request
			ctx, span := tracing.StartSpan(
				r.Context(),
				"http-server",
				r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Add request attributes to the span
			span.SetAttributes(
				tracing.AttrHTTPMethod.String(r.Method),
				tracing.AttrHTTPRoute.String(r.URL.Path),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("net.peer.ip", r.RemoteAddr),
			)

			// Wrap response writer to capture status code
			wrapped := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler with the traced context
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Add response status to span
			span.SetAttributes(tracing.AttrHTTPStatus.Int(wrapped.statusCode))

			// Set span status based on HTTP status code
			if wrapped.statusCode >= 400 {
				if wrapped.statusCode >= 500 {
					span.SetStatus(codes.Error, http.StatusText(wrapped.statusCode))
				} else {
					span.SetStatus(codes.Error, http.StatusText(wrapped.statusCode))
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}
