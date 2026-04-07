package advanced_security

import (
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RequestFingerprinting analyzes request patterns for DDoS detection
type RequestFingerprinting struct {
	mu           sync.RWMutex
	fingerprints map[string]*RequestPattern
	window       time.Duration
}

// BotDetection identifies bot traffic
type BotDetection struct {
	mu             sync.RWMutex
	botSignatures  map[string]bool
	suspiciousIPs  map[string]*BotActivity
	detectionRules []BotDetectionRule
	rateWindows    map[string][]time.Time // Track request timestamps for rate detection
	rateLimit      int                    // Max requests allowed in the rate window
	rateWindow     time.Duration          // Time window for rate detection
}

// TrafficAnalyzer monitors traffic patterns
type TrafficAnalyzer struct {
	mu               sync.RWMutex
	trafficStats     map[string]*TrafficStats
	window           time.Duration
	anomalyThreshold float64
}

// RequestFingerprinting implementation
func (rf *RequestFingerprinting) AnalyzeRequest(r *http.Request) {
	ip := getClientIP(r)

	rf.mu.Lock()
	defer rf.mu.Unlock()

	pattern, exists := rf.fingerprints[ip]
	if !exists {
		pattern = &RequestPattern{
			ip:         ip,
			requests:   make([]time.Time, 0),
			userAgents: make(map[string]int),
			paths:      make(map[string]int),
			methods:    make(map[string]int),
		}
		rf.fingerprints[ip] = pattern
	}

	now := time.Now()
	pattern.requests = append(pattern.requests, now)
	pattern.userAgents[r.Header.Get("User-Agent")]++
	pattern.paths[r.URL.Path]++
	pattern.methods[r.Method]++
	pattern.lastSeen = now

	// Clean old requests
	windowStart := now.Add(-rf.window)
	validRequests := make([]time.Time, 0)
	for _, ts := range pattern.requests {
		if ts.After(windowStart) {
			validRequests = append(validRequests, ts)
		}
	}
	pattern.requests = validRequests

	// Calculate suspicious score
	pattern.suspiciousScore = rf.calculateSuspiciousScore(pattern)
}

// BotDetection implementation
func (bd *BotDetection) DetectBot(r *http.Request) (bool, string) {
	userAgent := r.Header.Get("User-Agent")
	ip := getClientIP(r)

	// Check known bot signatures
	if bd.botSignatures[userAgent] {
		return true, "known_bot_signature"
	}

	// Check suspicious IP activity (with score decay)
	if activity, exists := bd.suspiciousIPs[ip]; exists {
		// Apply score decay based on time since last activity
		if time.Since(activity.lastActivity) > 5*time.Minute {
			activity.score = activity.score / 2 // Reduce score by half after 5 minutes
		}
		activity.lastActivity = time.Now()

		if activity.score > 50 {
			return true, activity.detectionReason
		}
	}

	// Apply detection rules (excluding the broken rapid_requests rule)
	for _, rule := range bd.detectionRules {
		// Skip rules with catch-all patterns - they're handled by rate detection
		if rule.pattern.String() == ".*" {
			continue
		}
		if rule.pattern.MatchString(userAgent) || rule.pattern.MatchString(r.URL.Path) {
			bd.markSuspiciousIP(ip, rule.score, rule.description)
			if activity := bd.suspiciousIPs[ip]; activity.score > 50 {
				return true, rule.description
			}
		}
	}

	// Check rate-based bot detection
	if bd.isRateExceeded(ip) {
		bd.markSuspiciousIP(ip, 15, "rapid_requests_rate_exceeded")
		if activity := bd.suspiciousIPs[ip]; activity.score > 50 {
			return true, "rapid_requests_rate_exceeded"
		}
	}

	return false, ""
}

// isRateExceeded checks if an IP has exceeded the rate limit
func (bd *BotDetection) isRateExceeded(ip string) bool {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-bd.rateWindow)

	// Get or create rate window for this IP
	timestamps, exists := bd.rateWindows[ip]
	if !exists {
		bd.rateWindows[ip] = []time.Time{now}
		return false
	}

	// Filter out old timestamps outside the window
	validTimestamps := make([]time.Time, 0)
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Add current request
	validTimestamps = append(validTimestamps, now)
	bd.rateWindows[ip] = validTimestamps

	// Check if rate limit exceeded
	return len(validTimestamps) > bd.rateLimit
}

// CleanupOldData removes stale rate tracking data to prevent memory leaks
func (bd *BotDetection) CleanupOldData() {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-10 * time.Minute)

	// Clean old rate windows
	for ip, timestamps := range bd.rateWindows {
		validTimestamps := make([]time.Time, 0)
		for _, ts := range timestamps {
			if ts.After(cutoff) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		if len(validTimestamps) == 0 {
			delete(bd.rateWindows, ip)
		} else {
			bd.rateWindows[ip] = validTimestamps
		}
	}

	// Clean old suspicious IPs with no recent activity and low scores
	for ip, activity := range bd.suspiciousIPs {
		if time.Since(activity.lastActivity) > 30*time.Minute && activity.score < 30 {
			delete(bd.suspiciousIPs, ip)
		}
	}
}

func (bd *BotDetection) markSuspiciousIP(ip string, score int, reason string) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	activity, exists := bd.suspiciousIPs[ip]
	if !exists {
		activity = &BotActivity{
			ip: ip,
		}
		bd.suspiciousIPs[ip] = activity
	}

	activity.score += score
	activity.lastActivity = time.Now()
	activity.detectionReason = reason

	// Auto-block if score is too high
	if activity.score > 100 {
		blockUntil := time.Now().Add(24 * time.Hour)
		activity.blockedUntil = &blockUntil
	}
}

// TrafficAnalyzer implementation
func (ta *TrafficAnalyzer) DetectAttack(ip string) (bool, string) {
	ta.mu.RLock()
	stats, exists := ta.trafficStats[ip]
	ta.mu.RUnlock()

	if !exists {
		return false, ""
	}

	// Check for anomaly patterns
	if stats.anomalyScore > ta.anomalyThreshold {
		return true, "traffic_anomaly"
	}

	// Check for error rate spikes
	if stats.errorCount > stats.requestCount/2 {
		return true, "high_error_rate"
	}

	return false, ""
}

func (ta *TrafficAnalyzer) monitorTraffic() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ta.analyzeTrafficPatterns()
	}
}

func (ta *TrafficAnalyzer) analyzeTrafficPatterns() {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	now := time.Now()

	// Analyze each IP's traffic pattern
	for ip, stats := range ta.trafficStats {
		// Remove old stats outside the analysis window
		if now.Sub(stats.lastRequest) > ta.window {
			delete(ta.trafficStats, ip)
			continue
		}

		// Calculate metrics
		anomalyScore := ta.calculateAnomalyScore(stats)

		// Detect traffic spikes (sudden increase in requests)
		spikeDetected := ta.detectTrafficSpike(stats)

		// Update stats
		stats.anomalyScore = anomalyScore
		stats.spikeDetected = spikeDetected

		// Clean up very old entries periodically
		if len(ta.trafficStats) > 10000 { // Prevent memory issues
			// Remove entries older than 2x window
			oldWindow := now.Add(-ta.window * 2)
			for oldIP, oldStats := range ta.trafficStats {
				if oldStats.lastRequest.Before(oldWindow) {
					delete(ta.trafficStats, oldIP)
				}
			}
		}
	}
}

func (ta *TrafficAnalyzer) calculateAnomalyScore(stats *TrafficStats) float64 {
	score := 0.0

	// High error rate anomaly
	if stats.requestCount > 0 {
		errorRate := float64(stats.errorCount) / float64(stats.requestCount)
		if errorRate > 0.5 { // More than 50% errors
			score += 2.0
		} else if errorRate > 0.2 { // More than 20% errors
			score += 1.0
		}
	}

	// High request frequency anomaly
	if stats.requestCount > 1000 { // Very high request count
		score += 2.0
	} else if stats.requestCount > 500 {
		score += 1.0
	}

	// Slow response times (if tracked)
	if stats.avgResponseTime > time.Second*5 {
		score += 1.5
	} else if stats.avgResponseTime > time.Second*2 {
		score += 0.5
	}

	// Spike detection contributes to anomaly score
	if stats.spikeDetected {
		score += 1.0
	}

	return score
}

func (ta *TrafficAnalyzer) detectTrafficSpike(stats *TrafficStats) bool {
	// Spike detection using statistical comparison against historical baseline
	if stats.baselineAvg == 0 || stats.baselineStdDev == 0 {
		// No baseline available, fall back to simple threshold
		return stats.requestCount > 100
	}

	// Calculate how many standard deviations above baseline
	currentRate := float64(stats.requestCount)
	deviations := (currentRate - stats.baselineAvg) / stats.baselineStdDev

	// Consider it a spike if it's more than 3 standard deviations above baseline
	// or if it's extremely high even with baseline (failsafe)
	return deviations > 3.0 || currentRate > stats.baselineAvg*5
}

// RecordRequest records a request for traffic analysis
func (ta *TrafficAnalyzer) RecordRequest(ip string, responseTime time.Duration, isError bool) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	now := time.Now()
	stats, exists := ta.trafficStats[ip]
	if !exists {
		stats = &TrafficStats{
			ip:                   ip,
			historicalRequests:   make([]int, 0, 24), // Keep 24 time windows
			historicalTimestamps: make([]time.Time, 0, 24),
		}
		ta.trafficStats[ip] = stats
	}

	// Update current statistics
	stats.requestCount++
	stats.lastRequest = now

	if isError {
		stats.errorCount++
	}

	// Update average response time (simple moving average)
	if stats.requestCount == 1 {
		stats.avgResponseTime = responseTime
	} else {
		// Weighted average favoring recent requests
		weight := 0.1
		stats.avgResponseTime = time.Duration(
			float64(stats.avgResponseTime)*(1-weight) + float64(responseTime)*weight,
		)
	}

	// Update historical data for baseline calculation
	ta.updateHistoricalData(stats, now)

	// Recalculate baseline periodically (every hour)
	if now.Sub(stats.lastBaselineUpdate) > time.Hour {
		ta.calculateBaseline(stats)
	}
}

// updateHistoricalData maintains rolling window of request counts
func (ta *TrafficAnalyzer) updateHistoricalData(stats *TrafficStats, now time.Time) {
	const windowSize = time.Hour // 1-hour windows

	// Clean old historical data (keep last 24 hours)
	cutoff := now.Add(-24 * time.Hour)
	validIndices := make([]int, 0)
	for i, timestamp := range stats.historicalTimestamps {
		if timestamp.After(cutoff) {
			validIndices = append(validIndices, i)
		}
	}

	if len(validIndices) != len(stats.historicalTimestamps) {
		newRequests := make([]int, 0, len(validIndices))
		newTimestamps := make([]time.Time, 0, len(validIndices))
		for _, idx := range validIndices {
			newRequests = append(newRequests, stats.historicalRequests[idx])
			newTimestamps = append(newTimestamps, stats.historicalTimestamps[idx])
		}
		stats.historicalRequests = newRequests
		stats.historicalTimestamps = newTimestamps
	}

	// Add new data point or update current window
	if len(stats.historicalTimestamps) == 0 ||
		now.Sub(stats.historicalTimestamps[len(stats.historicalTimestamps)-1]) >= windowSize {

		// Start new time window
		stats.historicalRequests = append(stats.historicalRequests, 1)
		stats.historicalTimestamps = append(stats.historicalTimestamps, now)
	} else {
		// Update current time window
		stats.historicalRequests[len(stats.historicalRequests)-1]++
	}
}

// calculateBaseline computes statistical baseline from historical data
func (ta *TrafficAnalyzer) calculateBaseline(stats *TrafficStats) {
	if len(stats.historicalRequests) < 3 {
		// Need minimum data points for meaningful baseline
		stats.baselineAvg = 0
		stats.baselineStdDev = 0
		stats.lastBaselineUpdate = time.Now()
		return
	}

	// Calculate mean
	sum := 0
	for _, count := range stats.historicalRequests {
		sum += count
	}
	mean := float64(sum) / float64(len(stats.historicalRequests))

	// Calculate standard deviation
	sumSquares := 0.0
	for _, count := range stats.historicalRequests {
		diff := float64(count) - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(stats.historicalRequests)))

	stats.baselineAvg = mean
	stats.baselineStdDev = stdDev
	stats.lastBaselineUpdate = time.Now()
}

func (rf *RequestFingerprinting) calculateSuspiciousScore(pattern *RequestPattern) int {
	score := 0

	// High request frequency
	if len(pattern.requests) > 100 {
		score += 30
	}

	// Many different user agents
	if len(pattern.userAgents) > 5 {
		score += 20
	}

	// Unusual methods
	suspiciousMethods := map[string]bool{"TRACE": true, "TRACK": true, "CONNECT": true}
	for method := range pattern.methods {
		if suspiciousMethods[method] {
			score += 25
		}
	}

	return score
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for requests behind proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// isLoopbackIP returns true if the given address is a loopback address (e.g. 127.0.0.1, ::1).
// The ip string may be "host" or "host:port" (e.g. "[::1]:46098" or "127.0.0.1:12345").
func isLoopbackIP(ip string) bool {
	host := ip
	if hostStr, _, err := net.SplitHostPort(ip); err == nil {
		host = hostStr
	}
	parsed := net.ParseIP(host)
	return parsed != nil && parsed.IsLoopback()
}
