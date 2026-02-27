package advanced_security

import (
	"net/http"
	"regexp"
	"time"
)

// RequestPattern tracks request patterns from an IP
type RequestPattern struct {
	ip            string
	requests      []time.Time
	userAgents    map[string]int
	paths         map[string]int
	methods       map[string]int
	suspiciousScore int
	lastSeen      time.Time
	blocked       bool
}

// BotActivity tracks bot-like behavior
type BotActivity struct {
	ip              string
	score           int
	lastActivity    time.Time
	blockedUntil    *time.Time
	detectionReason string
}

// BotDetectionRule defines rules for bot detection
type BotDetectionRule struct {
	name        string
	pattern     *regexp.Regexp
	score       int
	description string
}

// TrafficStats holds traffic statistics for analysis
type TrafficStats struct {
	ip              string
	requestCount    int
	errorCount      int
	avgResponseTime time.Duration
	lastRequest     time.Time
	spikeDetected   bool
	anomalyScore    float64

	// Historical data for baseline comparison
	historicalRequests []int         // Request counts for recent time windows
	historicalTimestamps []time.Time // Timestamps for historical data
	baselineAvg      float64       // Calculated baseline average
	baselineStdDev   float64       // Standard deviation for statistical significance
	lastBaselineUpdate time.Time   // When baseline was last calculated
}

// QueuedRequest represents a queued request
type QueuedRequest struct {
	request  *http.Request
	response http.ResponseWriter
	handler  http.HandlerFunc
	done     chan bool
	started  time.Time
}

// responseWriterTracker wraps http.ResponseWriter to track responses
type responseWriterTracker struct {
	http.ResponseWriter
	statusCode int
	asm       *AdvancedSecurityMiddleware
	written   bool
	startTime time.Time
	clientIP  string
}