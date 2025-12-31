package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UsernameKey contextKey = "username"
const IsAnonymousKey contextKey = "is_anonymous"

func AuthMiddleware(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			claims, err := jwtManager.ValidateAccessToken(token)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, IsAnonymousKey, claims.IsAnonymous)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func GetUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}

func GetIsAnonymousFromContext(ctx context.Context) bool {
	if isAnonymous, ok := ctx.Value(IsAnonymousKey).(bool); ok {
		return isAnonymous
	}
	return false
}

// GetUserID is a convenience wrapper that returns (userID, exists)
func GetUserID(ctx context.Context) (string, bool) {
	userID := GetUserIDFromContext(ctx)
	return userID, userID != ""
}

// GetUsername is a convenience wrapper that returns (username, exists)
func GetUsername(ctx context.Context) (string, bool) {
	username := GetUsernameFromContext(ctx)
	return username, username != ""
}
