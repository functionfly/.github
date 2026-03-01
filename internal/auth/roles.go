package auth

// IsAdminRole returns true if the role is a platform admin role (entitled to enterprise plan display).
func IsAdminRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleSupport, RoleBillingAdmin, RoleDeveloperAdmin:
		return true
	default:
		return false
	}
}
