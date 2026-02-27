package advanced_security

import (
	"sync"
	"time"
)

// IPReputationChecker checks IP reputation
type IPReputationChecker struct {
	mu              sync.RWMutex
	reputationDB    map[string]int // IP -> reputation score
	cacheDuration   time.Duration
	lastUpdate      time.Time
}

// GeoBlocker blocks requests based on geographic location and IP reputation
type GeoBlocker struct {
	mu               sync.RWMutex
	blockedCountries map[string]bool
	blockedIPs       map[string]bool
	allowedIPs       map[string]bool
}

// GeoBlocker implementation
func (gb *GeoBlocker) IsAllowed(ip string) bool {
	gb.mu.RLock()
	defer gb.mu.RUnlock()

	// Check allowlist first
	if gb.allowedIPs[ip] {
		return true
	}

	// Check blocklist
	if gb.blockedIPs[ip] {
		return false
	}

	// Check country blocking (would need geo-IP database)
	// For now, just return true
	return true
}

// IPReputationChecker implementation
func (irc *IPReputationChecker) GetReputation(ip string) int {
	irc.mu.RLock()
	score, exists := irc.reputationDB[ip]
	irc.mu.RUnlock()

	if !exists {
		// Default neutral score
		return 0
	}

	return score
}