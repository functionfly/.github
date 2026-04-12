package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/security"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/sirupsen/logrus"
)

// AntiLoopMiddleware provides automated loop detection and prevention for agent operations
type AntiLoopMiddleware struct {
	securityService *security.SwarmSecurityService
	swarmService    *swarm.Service
}

// NewAntiLoopMiddleware creates a new anti-loop middleware
func NewAntiLoopMiddleware(securityService *security.SwarmSecurityService, swarmService *swarm.Service) *AntiLoopMiddleware {
	return &AntiLoopMiddleware{
		securityService: securityService,
		swarmService:    swarmService,
	}
}

// LoopPreventionConfig configures loop prevention thresholds
type LoopPreventionConfig struct {
	MaxDelegationDepth      int     // Maximum allowed delegation depth
	MaxMessageChainLength   int     // Maximum message forwarding chain
	MaxSpawnsPerMinute      int     // Maximum spawns per minute per parent
	MaxMessagesPerMinute    int     // Maximum messages per minute per agent
	AutoKillEnabled         bool    // Whether to auto-trigger kill switch
	BudgetKillThreshold     float64 // Auto-kill if wallet below this
	AnomalyAutoKillSeverity string  // Auto-kill on this severity (high/medium)
}

// DefaultLoopPreventionConfig returns default configuration
func DefaultLoopPreventionConfig() *LoopPreventionConfig {
	return &LoopPreventionConfig{
		MaxDelegationDepth:      10,
		MaxMessageChainLength:   20,
		MaxSpawnsPerMinute:      5,
		MaxMessagesPerMinute:    100,
		AutoKillEnabled:         true,
		BudgetKillThreshold:     0.01, // $0.01 USD
		AnomalyAutoKillSeverity: "high",
	}
}

// LoopDetectionResult contains the result of loop detection
type LoopDetectionResult struct {
	HasLoop        bool     `json:"has_loop"`
	LoopType       string   `json:"loop_type,omitempty"`
	Violations     []string `json:"violations,omitempty"`
	Severity       string   `json:"severity"` // critical | high | medium | low
	Recommendation string   `json:"recommendation,omitempty"`
}

// CheckSpawnLoop checks if spawning a child would create a loop
func (m *AntiLoopMiddleware) CheckSpawnLoop(ctx context.Context, parentAgentID string, childAgentID string) *LoopDetectionResult {
	result := &LoopDetectionResult{
		HasLoop:    false,
		Violations: []string{},
		Severity:   "low",
	}

	// 1. Check if parent exists and is active
	validation := m.securityService.ValidateSpawnRequest(ctx, parentAgentID, childAgentID)
	if !validation.Valid {
		result.HasLoop = true
		result.Severity = "high"
		result.LoopType = "spawn_validation_failed"
		result.Violations = validation.Reasons
		result.Recommendation = "Review agent status and spawn limits"
		return result
	}

	// 2. Check for cycle (agent spawning itself or its ancestor)
	if m.securityService.HasCycle(ctx, parentAgentID, childAgentID) {
		result.HasLoop = true
		result.Severity = "critical"
		result.LoopType = "circular_spawn"
		result.Violations = append(result.Violations, "Spawn would create circular dependency")
		result.Recommendation = "Cannot spawn agent that would create cycle"
	}

	// 3. Check delegation depth
	depth, _ := m.swarmService.GetDelegationDepth(ctx, parentAgentID)
	config := DefaultLoopPreventionConfig()
	if depth >= config.MaxDelegationDepth {
		result.HasLoop = true
		result.Severity = "high"
		result.LoopType = "max_depth_exceeded"
		result.Violations = append(result.Violations, fmt.Sprintf("Maximum delegation depth (%d) reached", config.MaxDelegationDepth))
		result.Recommendation = "Use flatter hierarchy or infrastructure agents"
	}

	return result
}

// CheckMessageLoop checks if sending a message would create a loop
func (m *AntiLoopMiddleware) CheckMessageLoop(ctx context.Context, fromAgentID string, toAgentID string, chainLength int) *LoopDetectionResult {
	result := &LoopDetectionResult{
		HasLoop:    false,
		Violations: []string{},
		Severity:   "low",
	}

	config := DefaultLoopPreventionConfig()

	// 1. Check message chain length
	if chainLength >= config.MaxMessageChainLength {
		result.HasLoop = true
		result.Severity = "high"
		result.LoopType = "message_chain_too_long"
		result.Violations = append(result.Violations, fmt.Sprintf("Message chain length (%d) exceeds maximum (%d)", chainLength, config.MaxMessageChainLength))
		result.Recommendation = "Break long message chains into batch operations"
	}

	// 2. Check for circular messaging (using security service cycle detection)
	if m.securityService.HasCycle(ctx, fromAgentID, toAgentID) {
		result.HasLoop = true
		result.Severity = "critical"
		result.LoopType = "circular_messaging"
		result.Violations = append(result.Violations, "Message would create circular dependency")
		result.Recommendation = "Review agent communication topology"
	}

	// 3. Check delegation depth for the sender
	depth, _ := m.swarmService.GetDelegationDepth(ctx, fromAgentID)
	if depth >= config.MaxDelegationDepth {
		result.HasLoop = true
		if result.Severity != "critical" {
			result.Severity = "medium"
		}
		result.LoopType = "deep_hierarchy_messaging"
		result.Violations = append(result.Violations, fmt.Sprintf("Sender at deep hierarchy level (%d)", depth))
		result.Recommendation = "Consider infrastructure agent for deep delegation"
	}

	return result
}

// AutoEnforcer automatically enforces loop prevention policies
type AutoEnforcer struct {
	middleware      *AntiLoopMiddleware
	config          *LoopPreventionConfig
	killSwitchFn    func(ctx context.Context, agentID string, reason string) error
	lastSpawnCounts map[string]*rateCount
	lastMsgCounts   map[string]*rateCount
	mu              chan struct{} // Simple semaphore for rate counting
}

type rateCount struct {
	count       int
	windowStart time.Time
}

// NewAutoEnforcer creates a new automatic loop enforcer
func NewAutoEnforcer(middleware *AntiLoopMiddleware, config *LoopPreventionConfig, killSwitchFn func(ctx context.Context, agentID string, reason string) error) *AutoEnforcer {
	return &AutoEnforcer{
		middleware:      middleware,
		config:          config,
		killSwitchFn:    killSwitchFn,
		lastSpawnCounts: make(map[string]*rateCount),
		lastMsgCounts:   make(map[string]*rateCount),
		mu:              make(chan struct{}, 1),
	}
}

// EnforceSpawn enforces spawn limits and auto-kills if needed
func (e *AutoEnforcer) EnforceSpawn(ctx context.Context, parentAgentID string, childAgentID string) (*LoopDetectionResult, error) {
	// Check for loops first
	result := e.middleware.CheckSpawnLoop(ctx, parentAgentID, childAgentID)

	if result.HasLoop && result.Severity == "critical" {
		// Auto-kill for critical violations
		if e.config.AutoKillEnabled && e.killSwitchFn != nil {
			logrus.Warnf("AntiLoop: Auto-killing agent %s due to critical spawn loop: %s", parentAgentID, result.LoopType)
			if err := e.killSwitchFn(ctx, parentAgentID, fmt.Sprintf("anti-loop: %s", result.LoopType)); err != nil {
				logrus.WithError(err).Errorf("AntiLoop: Failed to auto-kill agent %s", parentAgentID)
			}
		}
		return result, fmt.Errorf("spawn blocked: %s", strings.Join(result.Violations, "; "))
	}

	// Check rate limit
	e.mu <- struct{}{}
	defer func() { <-e.mu }()

	now := time.Now()
	count := e.lastSpawnCounts[parentAgentID]
	if count == nil || now.Sub(count.windowStart) > time.Minute {
		count = &rateCount{count: 0, windowStart: now}
		e.lastSpawnCounts[parentAgentID] = count
	}
	count.count++

	if count.count > e.config.MaxSpawnsPerMinute {
		result.HasLoop = true
		result.Severity = "high"
		result.LoopType = "spawn_rate_limit"
		result.Violations = append(result.Violations, fmt.Sprintf("Spawn rate limit exceeded: %d per minute", e.config.MaxSpawnsPerMinute))

		if e.config.AutoKillEnabled && e.killSwitchFn != nil {
			logrus.Warnf("AntiLoop: Auto-killing agent %s due to spawn spam", parentAgentID)
			if err := e.killSwitchFn(ctx, parentAgentID, "anti-loop: spawn rate limit exceeded"); err != nil {
				logrus.WithError(err).Errorf("AntiLoop: Failed to auto-kill agent %s", parentAgentID)
			}
		}
		return result, fmt.Errorf("spawn rate limit exceeded")
	}

	return result, nil
}

// EnforceMessage enforces message limits and checks for loops
func (e *AutoEnforcer) EnforceMessage(ctx context.Context, fromAgentID string, toAgentID string, chainLength int) (*LoopDetectionResult, error) {
	result := e.middleware.CheckMessageLoop(ctx, fromAgentID, toAgentID, chainLength)

	if result.HasLoop && result.Severity == "critical" {
		if e.config.AutoKillEnabled && e.killSwitchFn != nil {
			logrus.Warnf("AntiLoop: Auto-killing agent %s due to critical message loop", fromAgentID)
			e.killSwitchFn(ctx, fromAgentID, fmt.Sprintf("anti-loop: %s", result.LoopType))
		}
		return result, fmt.Errorf("message blocked: %s", strings.Join(result.Violations, "; "))
	}

	// Check message rate
	e.mu <- struct{}{}
	defer func() { <-e.mu }()

	now := time.Now()
	count := e.lastMsgCounts[fromAgentID]
	if count == nil || now.Sub(count.windowStart) > time.Minute {
		count = &rateCount{count: 0, windowStart: now}
		e.lastMsgCounts[fromAgentID] = count
	}
	count.count++

	if count.count > e.config.MaxMessagesPerMinute {
		result.HasLoop = true
		result.Severity = "medium"
		result.LoopType = "message_rate_limit"
		result.Violations = append(result.Violations, fmt.Sprintf("Message rate limit exceeded: %d per minute", e.config.MaxMessagesPerMinute))
		return result, fmt.Errorf("message rate limit exceeded")
	}

	return result, nil
}

// BudgetKillSwitchMonitor monitors agent budgets and triggers kill switch when depleted
type BudgetKillSwitchMonitor struct {
	securityService *security.SwarmSecurityService
	walletGetter    func(ctx context.Context, agentID string) (float64, error)
	config          *LoopPreventionConfig
	killSwitchFn    func(ctx context.Context, agentID string, reason string) error
}

// NewBudgetKillSwitchMonitor creates a budget-based kill switch monitor
func NewBudgetKillSwitchMonitor(securityService *security.SwarmSecurityService, walletGetter func(ctx context.Context, agentID string) (float64, error), config *LoopPreventionConfig, killSwitchFn func(ctx context.Context, agentID string, reason string) error) *BudgetKillSwitchMonitor {
	return &BudgetKillSwitchMonitor{
		securityService: securityService,
		walletGetter:    walletGetter,
		config:          config,
		killSwitchFn:    killSwitchFn,
	}
}

// CheckBudget checks an agent's budget and triggers kill switch if depleted
func (m *BudgetKillSwitchMonitor) CheckBudget(ctx context.Context, agentID string) error {
	if !m.config.AutoKillEnabled {
		return nil
	}

	balance, err := m.walletGetter(ctx, agentID)
	if err != nil {
		logrus.WithError(err).Warnf("BudgetKillSwitch: Failed to get wallet for agent %s", agentID)
		return nil // Don't kill if we can't check
	}

	if balance <= m.config.BudgetKillThreshold {
		logrus.Warnf("BudgetKillSwitch: Agent %s budget depleted ($%.4f), triggering kill switch", agentID, balance)
		if err := m.killSwitchFn(ctx, agentID, fmt.Sprintf("budget_depleted: $%.4f", balance)); err != nil {
			logrus.WithError(err).Errorf("BudgetKillSwitch: Failed to kill agent %s", agentID)
			return err
		}
	}

	return nil
}

// CheckAnomalies checks for anomalies and auto-kills on high severity
func (m *BudgetKillSwitchMonitor) CheckAnomalies(ctx context.Context, agentID string) error {
	if !m.config.AutoKillEnabled {
		return nil
	}

	anomalies, err := m.securityService.DetectAnomaly(ctx, agentID, 5*time.Minute)
	if err != nil {
		logrus.WithError(err).Warnf("BudgetKillSwitch: Failed to detect anomalies for agent %s", agentID)
		return nil
	}

	hasCritical := false
	for _, a := range anomalies {
		if a.Severity == "high" && m.config.AnomalyAutoKillSeverity == "high" {
			hasCritical = true
			break
		}
		if a.Severity == "high" || (a.Severity == "medium" && m.config.AnomalyAutoKillSeverity == "medium") {
			hasCritical = true
			break
		}
	}

	if hasCritical {
		logrus.Warnf("BudgetKillSwitch: Agent %s has critical anomalies, triggering kill switch", agentID)
		if err := m.killSwitchFn(ctx, agentID, "anomaly_detected: critical"); err != nil {
			logrus.WithError(err).Errorf("BudgetKillSwitch: Failed to kill agent %s", agentID)
			return err
		}
	}

	return nil
}

// HTTPMiddleware wraps an HTTP handler with loop detection
func (m *AntiLoopMiddleware) HTTPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract agent IDs from path if present
		path := r.URL.Path
		var agentID string

		// Try to extract agent ID from common patterns
		if idx := strings.Index(path, "/agent/"); idx != -1 {
			remainder := path[idx+7:]
			if end := strings.IndexAny(remainder, "/?#"); end != -1 {
				agentID = remainder[:end]
			} else {
				agentID = remainder
			}
		}

		// Log loop detection for this request
		if agentID != "" {
			ctx := r.Context()
			anomalies, _ := m.securityService.DetectAnomaly(ctx, agentID, time.Minute)
			if len(anomalies) > 0 {
				for _, a := range anomalies {
					if a.Severity == "high" {
						logrus.WithFields(logrus.Fields{
							"agent_id": agentID,
							"path":     path,
							"anomaly":  a.Type,
						}).Warn("AntiLoop: High severity anomaly detected")
					}
				}
			}
		}

		next(w, r)
	}
}

// ParseChainLength extracts chain length from request body or headers
func ParseChainLength(r *http.Request) int {
	// Check header first
	if h := r.Header.Get("X-Message-Chain-Length"); h != "" {
		if n, err := strconv.Atoi(h); err == nil {
			return n
		}
	}
	return 0
}
