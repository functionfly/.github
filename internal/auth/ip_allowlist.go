package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// IPAllowlistService handles IP allowlist checking logic
type IPAllowlistService struct {
	repo *storage.IPAllowlistRepository
}

// NewIPAllowlistService creates a new IP allowlist service
func NewIPAllowlistService(repo *storage.IPAllowlistRepository) *IPAllowlistService {
	return &IPAllowlistService{repo: repo}
}

// CheckAccessResult represents the result of an IP allowlist check
type CheckAccessResult struct {
	Allowed     bool
	MFARequired bool
}

// CheckAccess checks if the given client IP is allowed for the tenant
// Returns:
// - allowed: true if access is granted
// - mfaRequired: true if MFA verification could grant access to unknown IPs
// - err: any error that occurred
func (s *IPAllowlistService) CheckAccess(ctx context.Context, tenantID uuid.UUID, clientIP string) (allowed bool, mfaRequired bool, err error) {
	// Get the allowlist for the tenant
	allowlist, err := s.repo.GetAllowlistByTenantID(tenantID)
	if err != nil {
		return false, false, fmt.Errorf("failed to get IP allowlist: %w", err)
	}

	// If no allowlist is configured, allow access
	if allowlist == nil {
		return true, false, nil
	}

	// Get all entries for the allowlist
	entries, err := s.repo.GetEntriesByAllowlistID(allowlist.ID)
	if err != nil {
		return false, false, fmt.Errorf("failed to get IP allowlist entries: %w", err)
	}

	// Normalize and parse the client IP
	clientIP = normalizeIP(clientIP)
	parsedClientIP := net.ParseIP(clientIP)
	if parsedClientIP == nil {
		// Invalid IP format - apply default policy
		return allowlist.DefaultPolicy == "allow", allowlist.MFARequiredForUnknownIP, nil
	}

	// Check if the client IP matches any entry
	matched := false
	for _, entry := range entries {
		if s.matchIP(parsedClientIP, entry.Type, entry.Value) {
			matched = true
			break
		}
	}

	// Apply the default policy based on whether IP matched
	if matched {
		// IP is in the allowlist - apply allow policy
		return allowlist.DefaultPolicy == "allow", false, nil
	}

	// IP is NOT in the allowlist
	// If default policy is allow, grant access
	if allowlist.DefaultPolicy == "allow" {
		return true, false, nil
	}

	// Default policy is deny - check if MFA can grant access
	return false, allowlist.MFARequiredForUnknownIP, nil
}

// matchIP checks if the given IP matches the entry type and value
// entryType can be "ip" for a single IP or "cidr" for a CIDR range
func (s *IPAllowlistService) matchIP(clientIP net.IP, entryType, entryValue string) bool {
	switch entryType {
	case "ip":
		// Single IP address comparison
		return matchSingleIP(clientIP, entryValue)
	case "cidr":
		// CIDR range comparison
		return matchCIDR(clientIP, entryValue)
	default:
		// Unknown type - assume no match
		return false
	}
}

// matchSingleIP checks if a client IP matches a single IP address
func matchSingleIP(clientIP net.IP, ipValue string) bool {
	parsedIP := net.ParseIP(ipValue)
	if parsedIP == nil {
		return false
	}
	return clientIP.Equal(parsedIP)
}

// matchCIDR checks if a client IP is within a CIDR range
func matchCIDR(clientIP net.IP, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(clientIP)
}

// normalizeIP normalizes an IP address string by removing any port or zone information
func normalizeIP(ip string) string {
	// Handle IPv4 with port (e.g., "192.168.1.1:8080")
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// Check if it's IPv6 (has multiple colons)
		if strings.Count(ip, ":") > 1 {
			// IPv6 - could be "[2001:db8::1]:8080" or "2001:db8::1%eth0"
			ip = strings.TrimPrefix(ip, "[")
			if idx := strings.Index(ip, "]"); idx != -1 {
				ip = ip[:idx]
			}
		} else {
			// IPv4 with port
			ip = ip[:idx]
		}
	}

	// Handle IPv6 with zone (e.g., "fe80::1%eth0")
	if idx := strings.Index(ip, "%"); idx != -1 {
		ip = ip[:idx]
	}

	return strings.TrimSpace(ip)
}

// GetClientIP extracts the client IP from an HTTP request
// It checks X-Forwarded-For, X-Real-IP headers and falls back to RemoteAddr
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (may contain multiple IPs)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first is the original client
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return normalizeIP(strings.TrimSpace(ips[0]))
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return normalizeIP(strings.TrimSpace(xri))
	}

	// Fall back to RemoteAddr
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return normalizeIP(remoteAddr[:idx])
	}

	return remoteAddr
}

// IsMFAVerified checks if the request has been verified with MFA
// This checks for a session or context flag indicating MFA verification
func IsMFAVerified(r *http.Request) bool {
	// Check for MFA verification header (set after successful MFA challenge)
	mfaVerified := r.Header.Get("X-MFA-Verified")
	if mfaVerified == "true" {
		return true
	}

	// Check for MFA session cookie
	cookie, err := r.Cookie("mfa_verified")
	if err == nil && cookie.Value == "true" {
		return true
	}

	return false
}
