package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
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
		// SECURITY: Validate token version for revocation support
		// If the token has a TokenVersion, verify it matches the user's current version
		// This allows invalidating tokens when password is changed or logout all is triggered
		if claims.TokenVersion > 0 && a.repo != nil {
			user, err := a.repo.GetUserByID(claims.UserID)
			if err != nil {
				// Database error - fail closed for security
				// An attacker could exploit a DB outage to use revoked tokens
				logrus.WithError(err).WithField("user_id", claims.UserID).Error("Could not verify token version - database error, rejecting token for security")
				return nil, fmt.Errorf("token verification unavailable")
			}
			if user == nil {
				// User deleted - reject token (already handled by getting nil user)
				return nil, fmt.Errorf("user not found")
			}
			if user.TokenVersion > 0 && claims.TokenVersion != user.TokenVersion {
				logrus.WithFields(logrus.Fields{
					"user_id":               claims.UserID,
					"token_token_version":   claims.TokenVersion,
					"current_token_version": user.TokenVersion,
				}).Warn("Token version mismatch - token has been revoked")
				return nil, fmt.Errorf("token has been revoked")
			}
		}
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
	if user.Username != nil && *user.Username != "" {
		claims.Username = *user.Username
	}

	// Add permissions based on role
	if user.Role != "" {
		claims.Permissions = a.getPermissionsForRole(user.Role)
	}

	// SECURITY FIX: Add token version for revocation support
	// TokenVersion is stored in the user record and incremented on password change/logout all
	// This allows invalidating all existing tokens for a user without invalidating all users
	if user.TokenVersion > 0 {
		claims.TokenVersion = user.TokenVersion
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// generateRefreshToken creates a cryptographically secure refresh token
func (a *AuthService) generateRefreshToken() (token, hash string, err error) {
	// Generate a 64-byte random token (512 bits of entropy)
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	token = hex.EncodeToString(tokenBytes)
	hash = storage.HashRefreshToken(token)
	return token, hash, nil
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
