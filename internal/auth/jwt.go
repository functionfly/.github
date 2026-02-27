package auth

import (
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/golang-jwt/jwt/v5"
)

// ValidateToken validates a JWT token and returns the claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Create a parser with validation options
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(Issuer),
	)

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateToken creates a new JWT token for a user
func (a *AuthService) generateToken(user *storage.User) (string, error) {
	if len(a.jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not configured")
	}
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Email:    user.Email,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.jwtDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Add permissions based on role
	if user.Role != "" {
		claims.Permissions = a.getPermissionsForRole(user.Role)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// getPermissionsForRole returns the permissions for a given role
func (a *AuthService) getPermissionsForRole(role string) []string {
	switch role {
	case RoleSuperAdmin, RoleAdmin:
		return []string{
			PermTenantsRead, PermTenantsWrite,
			PermUsersRead, PermUsersWrite,
			PermBillingRead, PermBillingWrite,
			PermDeploymentsRead, PermDeploymentsWrite,
			PermAuditRead, PermSystemRead, PermSystemWrite,
			// StateFabric permissions
			PermStateRead, PermStateWrite, PermStateDelete, PermStateAdmin,
			PermTriggersManage,
			PermSnapshotsCreate, PermSnapshotsRestore,
			PermReplayAccess,
			PermMemoryRead, PermMemoryWrite,
			// Function Registry permissions
			PermRegistryPublish, PermRegistryVerify, PermRegistryApprove, PermRegistrySign, PermRegistryManage,
			// Monitoring permissions
			PermMonitoringAlerts, PermMonitoringManage, PermMonitoringMetrics, PermMonitoringAdmin, PermMonitoringHealth,
			// Security Operations permissions
			PermSecurityIncidents, PermSecurityScans, PermSecurityAudit, PermSecurityAdmin,
			// Content Management permissions
			PermContentCreate, PermContentEdit, PermContentPublish, PermContentManage, PermChangelogManage, PermBlogManage,
			// Team Management permissions
			PermTeamMembersManage, PermTeamRolesAssign, PermTeamResourcesShare,
			// Function Verification permissions
			PermVerificationApprove, PermVerificationSign, PermVerificationOverride,
			// Feedback Management permissions
			PermFeedbackModerate, PermFeedbackAnalytics,
		}
	case RoleSupport:
		return []string{
			PermTenantsRead,
			PermUsersRead,
			PermDeploymentsRead,
			PermSystemRead,
			// Monitoring permissions for support
			PermMonitoringAlerts, PermMonitoringMetrics, PermMonitoringHealth,
		}
	case RoleBillingAdmin:
		return []string{
			PermTenantsRead,
			PermUsersRead,
			PermBillingRead, PermBillingWrite,
		}
	case RoleDeveloperAdmin:
		return []string{
			PermTenantsRead,
			PermUsersRead,
			PermDeploymentsRead, PermDeploymentsWrite,
			// Function Registry permissions for developers
			PermRegistryPublish, PermRegistryVerify,
			// Function Verification permissions for developers
			PermVerificationSign,
		}
	case RoleReadOnly:
		return []string{
			PermTenantsRead,
			PermUsersRead,
			PermBillingRead,
			PermDeploymentsRead,
			PermAuditRead,
			PermSystemRead,
		}
	default:
		return []string{} // Regular users have no admin permissions
	}
}