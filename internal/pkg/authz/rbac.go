package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
)

const (
	// User permissions
	PermissionCreatePost     Permission = "post:create"
	PermissionReadPost       Permission = "post:read"
	PermissionUpdatePost     Permission = "post:update"
	PermissionCreateResponse Permission = "response:create"
	PermissionReadResponse   Permission = "response:read"

	// Circle permissions
	PermissionCreateCircle Permission = "circle:create"
	PermissionReadCircle   Permission = "circle:read"
	PermissionJoinCircle   Permission = "circle:join"
	PermissionLeaveCircle  Permission = "circle:leave"
	PermissionManageCircle Permission = "circle:manage"

	// Moderation permissions (overriding from authz.go)
	PermissionModerateContentExt Permission = "moderation:moderate_content"
	PermissionBanUserExt         Permission = "moderation:ban_user"
	PermissionUnbanUser          Permission = "moderation:unban_user"

	// Admin permissions
	PermissionManageUsers  Permission = "admin:manage_users"
	PermissionViewMetrics  Permission = "admin:view_metrics"
	PermissionManageSystem Permission = "admin:manage_system"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[domain.Role][]Permission{
	domain.RoleUser: {
		PermissionCreatePost,
		PermissionReadPost,
		PermissionUpdatePost,
		PermissionDeletePost,
		PermissionCreateResponse,
		PermissionReadResponse,
		PermissionCreateCircle,
		PermissionReadCircle,
		PermissionJoinCircle,
		PermissionLeaveCircle,
	},
	domain.RoleModerator: {
		// Moderators have all user permissions plus moderation permissions
		PermissionCreatePost,
		PermissionReadPost,
		PermissionUpdatePost,
		PermissionDeletePost,
		PermissionCreateResponse,
		PermissionReadResponse,
		PermissionCreateCircle,
		PermissionReadCircle,
		PermissionJoinCircle,
		PermissionLeaveCircle,
		PermissionManageCircle,
		PermissionViewReports,
		PermissionModerateContent,
		PermissionBanUser,
		PermissionUnbanUser,
	},
	domain.RoleAdmin: {
		// Admins have all permissions
		PermissionCreatePost,
		PermissionReadPost,
		PermissionUpdatePost,
		PermissionDeletePost,
		PermissionCreateResponse,
		PermissionReadResponse,
		PermissionCreateCircle,
		PermissionReadCircle,
		PermissionJoinCircle,
		PermissionLeaveCircle,
		PermissionManageCircle,
		PermissionViewReports,
		PermissionModerateContent,
		PermissionBanUser,
		PermissionUnbanUser,
		PermissionManageUsers,
		PermissionViewMetrics,
		PermissionManageSystem,
	},
}

// Authorizer handles authorization checks
type Authorizer struct{}

// NewAuthorizer creates a new authorizer
func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

// HasPermission checks if a role has a specific permission
func (a *Authorizer) HasPermission(role domain.Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func (a *Authorizer) HasAnyPermission(role domain.Role, permissions ...Permission) bool {
	for _, permission := range permissions {
		if a.HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func (a *Authorizer) HasAllPermissions(role domain.Role, permissions ...Permission) bool {
	for _, permission := range permissions {
		if !a.HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// CanAccessResource checks if a user can access a specific resource
func (a *Authorizer) CanAccessResource(ctx context.Context, userID uuid.UUID, role domain.Role, resourceOwnerID uuid.UUID, permission Permission) error {
	// Check if the user has the permission
	if !a.HasPermission(role, permission) {
		return fmt.Errorf("user does not have permission: %s", permission)
	}

	// For update/delete operations, check ownership (unless user is moderator/admin)
	if permission == PermissionUpdatePost || permission == PermissionDeletePost {
		if role == domain.RoleUser && userID != resourceOwnerID {
			return fmt.Errorf("user can only modify their own resources")
		}
	}

	return nil
}

// RequirePermission returns an error if the role doesn't have the permission
func (a *Authorizer) RequirePermission(role domain.Role, permission Permission) error {
	if !a.HasPermission(role, permission) {
		return fmt.Errorf("permission denied: %s", permission)
	}
	return nil
}

// RequireRole returns an error if the role is not one of the allowed roles
func (a *Authorizer) RequireRole(role domain.Role, allowedRoles ...domain.Role) error {
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return nil
		}
	}
	return fmt.Errorf("role %s is not authorized", role)
}

// IsModerator checks if a role is moderator or admin
func (a *Authorizer) IsModerator(role domain.Role) bool {
	return role == domain.RoleModerator || role == domain.RoleAdmin
}

// IsAdmin checks if a role is admin
func (a *Authorizer) IsAdmin(role domain.Role) bool {
	return role == domain.RoleAdmin
}
