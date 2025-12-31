package middleware

import (
	"net/http"
	"time"

	"github.com/yourorg/anonymous-support/internal/repository/redis"
)

func RateLimitMiddleware(realtimeRepo *redis.RealtimeRepository, action string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			allowed, err := realtimeRepo.CheckRateLimit(r.Context(), userID, action, limit, window)
			if err != nil {
				http.Error(w, "rate limit check failed", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", window.String())
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
