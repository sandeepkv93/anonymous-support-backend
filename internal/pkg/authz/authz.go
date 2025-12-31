package authz

import (
	"github.com/yourorg/anonymous-support/internal/domain"
)

// Permission represents a specific action a user can perform
type Permission string

const (
	PermissionModerateContent Permission = "moderate_content"
	PermissionBanUser         Permission = "ban_user"
	PermissionDeletePost      Permission = "delete_post"
	PermissionViewReports     Permission = "view_reports"
	PermissionManageCircles   Permission = "manage_circles"
)

// rolePermissions maps roles to their allowed permissions
var rolePermissions = map[domain.Role][]Permission{
	domain.RoleUser: {},
	domain.RoleModerator: {
		PermissionModerateContent,
		PermissionViewReports,
		PermissionDeletePost,
	},
	domain.RoleAdmin: {
		PermissionModerateContent,
		PermissionBanUser,
		PermissionDeletePost,
		PermissionViewReports,
		PermissionManageCircles,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role domain.Role, permission Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsAdmin checks if a role is admin
func IsAdmin(role domain.Role) bool {
	return role == domain.RoleAdmin
}

// IsModerator checks if a role is moderator or higher
func IsModerator(role domain.Role) bool {
	return role == domain.RoleModerator || role == domain.RoleAdmin
}
