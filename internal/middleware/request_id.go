package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey string

const RequestIDKey requestIDKey = "request_id"

// RequestIDMiddleware adds a unique request ID to each request context
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists in header (for distributed tracing)
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
