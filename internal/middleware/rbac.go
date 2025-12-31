package middleware

import (
	"context"
	"net/http"

	"github.com/yourorg/anonymous-support/internal/domain"
)

// RBACMiddleware enforces role-based access control
func RBACMiddleware(requiredRole domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRoleFromContext(r.Context())
			if role == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has required role or higher
			if !hasPermission(domain.Role(role), requiredRole) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserRoleFromContext retrieves user role from context
func GetUserRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value("user_role").(string); ok {
		return role
	}
	return ""
}

// hasPermission checks if user role has permission for required role
func hasPermission(userRole, requiredRole domain.Role) bool {
	roleHierarchy := map[domain.Role]int{
		domain.RoleUser:      1,
		domain.RoleModerator: 2,
		domain.RoleAdmin:     3,
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}

// RequireAdmin is a convenience middleware for admin-only endpoints
func RequireAdmin() func(http.Handler) http.Handler {
	return RBACMiddleware(domain.RoleAdmin)
}

// RequireModerator is a convenience middleware for moderator+ endpoints
func RequireModerator() func(http.Handler) http.Handler {
	return RBACMiddleware(domain.RoleModerator)
}
