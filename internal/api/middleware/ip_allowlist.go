package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// IPAllowlistCacheTTL is the TTL for cached IP allowlist entries
	IPAllowlistCacheTTL = 5 * time.Minute
	// IPAllowlistCacheKeyPrefix is the Redis key prefix for IP allowlist caching
	IPAllowlistCacheKeyPrefix = "ipallowlist:cache:"
	// IPAllowlistBypassCacheKey is used to signal when to bypass cache
	IPAllowlistBypassCacheKey = "ipallowlist:bypass"
)

// IPAllowlistEntry represents an entry in the IP allowlist
type IPAllowlistEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CIDR        string `json:"cidr"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
	IsWhitelist bool   `json:"is_whitelist"`
}

// IPAllowlistMiddleware handles IP-based access control for admin routes
type IPAllowlistMiddleware struct {
	db          *sql.DB
	redisClient *redis.Client
	logger      *logrus.Entry
}

// NewIPAllowlistMiddleware creates a new IP allowlist middleware
func NewIPAllowlistMiddleware(db *sql.DB, redisClient *redis.Client) *IPAllowlistMiddleware {
	return &IPAllowlistMiddleware{
		db:          db,
		redisClient: redisClient,
		logger:      logrus.WithField("middleware", "ip_allowlist"),
	}
}

// isInternalIP checks if an IP address is considered internal/bypass
func isInternalIP(ip string) bool {
	// Check for localhost
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	// Parse the IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",  // IPv6 private
		"fe80::/10", // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// isIPInCIDR checks if an IP address is within a CIDR range
func isIPInCIDR(ip string, cidr string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		// Try parsing as a single IP (without mask)
		ip2 := net.ParseIP(cidr)
		if ip2 == nil {
			return false
		}
		return parsedIP.Equal(ip2)
	}

	return network.Contains(parsedIP)
}

// getCachedAllowlist retrieves the cached IP allowlist from Redis
func (m *IPAllowlistMiddleware) getCachedAllowlist(ctx context.Context) ([]IPAllowlistEntry, error) {
	if m.redisClient == nil {
		return nil, nil
	}

	key := IPAllowlistCacheKeyPrefix + "all"
	data, err := m.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		m.logger.WithError(err).Warn("Failed to get IP allowlist from cache")
		return nil, err
	}

	var entries []IPAllowlistEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		m.logger.WithError(err).Warn("Failed to unmarshal cached IP allowlist")
		return nil, err
	}

	return entries, nil
}

// cacheAllowlist caches the IP allowlist in Redis
func (m *IPAllowlistMiddleware) cacheAllowlist(ctx context.Context, entries []IPAllowlistEntry) error {
	if m.redisClient == nil {
		return nil
	}

	data, err := json.Marshal(entries)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to marshal IP allowlist for caching")
		return err
	}

	key := IPAllowlistCacheKeyPrefix + "all"
	if err := m.redisClient.Set(ctx, key, data, IPAllowlistCacheTTL).Err(); err != nil {
		m.logger.WithError(err).Warn("Failed to cache IP allowlist")
		return err
	}

	return nil
}

// invalidateCache invalidates the IP allowlist cache
func (m *IPAllowlistMiddleware) invalidateCache(ctx context.Context) error {
	if m.redisClient == nil {
		return nil
	}

	key := IPAllowlistCacheKeyPrefix + "all"
	if err := m.redisClient.Del(ctx, key).Err(); err != nil {
		m.logger.WithError(err).Warn("Failed to invalidate IP allowlist cache")
		return err
	}

	return nil
}

// loadAllowlistFromDB loads the IP allowlist from the database
func (m *IPAllowlistMiddleware) loadAllowlistFromDB(ctx context.Context) ([]IPAllowlistEntry, error) {
	if m.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, name, cidr, COALESCE(description, ''), is_active, is_whitelist
		FROM ip_allowlist
		WHERE is_active = TRUE
		ORDER BY created_at DESC`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		m.logger.WithError(err).Error("Failed to load IP allowlist from database")
		return nil, err
	}
	defer rows.Close()

	var entries []IPAllowlistEntry
	for rows.Next() {
		var entry IPAllowlistEntry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.CIDR, &entry.Description, &entry.IsActive, &entry.IsWhitelist); err != nil {
			m.logger.WithError(err).Warn("Failed to scan IP allowlist entry")
			continue
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		m.logger.WithError(err).Error("Error iterating IP allowlist rows")
		return nil, err
	}

	return entries, nil
}

// getAllowlist retrieves the IP allowlist, using cache when available
func (m *IPAllowlistMiddleware) getAllowlist(ctx context.Context) ([]IPAllowlistEntry, error) {
	// Try cache first
	entries, err := m.getCachedAllowlist(ctx)
	if err == nil && entries != nil {
		return entries, nil
	}

	// Load from database
	entries, err = m.loadAllowlistFromDB(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if entries == nil {
		entries = []IPAllowlistEntry{}
	}
	if err := m.cacheAllowlist(ctx, entries); err != nil {
		m.logger.WithError(err).Warn("Failed to cache IP allowlist")
	}

	return entries, nil
}

// IsIPAllowed checks if an IP address is allowed based on the IP allowlist
func (m *IPAllowlistMiddleware) IsIPAllowed(ctx context.Context, clientIP string) (bool, error) {
	// Skip check for internal IPs
	if isInternalIP(clientIP) {
		m.logger.WithField("ip", clientIP).Debug("IP allowlist check bypassed for internal IP")
		return true, nil
	}

	entries, err := m.getAllowlist(ctx)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get IP allowlist")
		// On error, fail open to avoid blocking legitimate users
		return true, nil
	}

	// If no entries, allow all
	if len(entries) == 0 {
		return true, nil
	}

	// Check each entry
	for _, entry := range entries {
		if !entry.IsActive {
			continue
		}

		if isIPInCIDR(clientIP, entry.CIDR) {
			m.logger.WithFields(logrus.Fields{
				"ip":           clientIP,
				"matched_cidr": entry.CIDR,
				"is_whitelist": entry.IsWhitelist,
				"entry_name":   entry.Name,
			}).Debug("IP matched allowlist entry")

			return entry.IsWhitelist, nil
		}
	}

	// Default deny if no whitelist match
	m.logger.WithFields(logrus.Fields{
		"ip":   clientIP,
		"mode": "default_deny",
	}).Debug("IP not matched by any allowlist entry, denying")

	return false, nil
}

// RequireIPAllowlist middleware that enforces IP-based access control
func (m *IPAllowlistMiddleware) RequireIPAllowlist(next http.HandlerFunc) http.HandlerFunc {
	return m.requireIPAllowlistInternal(next, false)
}

// RequireSuperAdminBypassIPAllowlist middleware that enforces IP-based access control but allows super_admin to bypass
func (m *IPAllowlistMiddleware) RequireSuperAdminBypassIPAllowlist(next http.HandlerFunc) http.HandlerFunc {
	return m.requireIPAllowlistInternal(next, true)
}

func (m *IPAllowlistMiddleware) requireIPAllowlistInternal(next http.HandlerFunc, allowSuperAdminBypass bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Skip if database is not configured (development mode)
		if m.db == nil {
			m.logger.Warn("IP allowlist check skipped - database not configured")
			next.ServeHTTP(w, r)
			return
		}

		// Extract client IP
		clientIP := extractClientIPFromRequest(r)

		// Check if IP is allowed
		allowed, err := m.IsIPAllowed(ctx, clientIP)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Error checking IP allowlist")
			// Fail open on error
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":   clientIP,
				"path": r.URL.Path,
				"user": r.Header.Get("X-User-ID"),
			}).Warn("IP blocked by allowlist policy")

			writeIPBlockedError(w, clientIP)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// ExtractClientIP extracts the client IP from the request
func ExtractClientIP(r *http.Request) string {
	return extractClientIPFromRequest(r)
}

// writeIPBlockedError writes a 403 Forbidden response for blocked IPs
func writeIPBlockedError(w http.ResponseWriter, clientIP string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	response := map[string]interface{}{
		"error":       "ip_blocked",
		"message":     "Access denied: your IP address is not allowed",
		"client_ip":   clientIP,
		"status_code": 403,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleInvalidateCache handles POST /internal/ip-allowlist/cache/invalidate
// This is called when the allowlist is modified
func (m *IPAllowlistMiddleware) HandleInvalidateCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := m.invalidateCache(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to invalidate IP allowlist cache")
		http.Error(w, "Failed to invalidate cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
