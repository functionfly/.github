package security

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// SwarmSecurityService handles security controls for agent swarms
type SwarmSecurityService struct {
	db *gorm.DB
}

// NewSwarmSecurityService creates a new swarm security service
func NewSwarmSecurityService(db *gorm.DB) *SwarmSecurityService {
	return &SwarmSecurityService{db: db}
}

// ValidateSpawnRequest validates a spawn request against security policies
func (s *SwarmSecurityService) ValidateSpawnRequest(ctx context.Context, parentAgentID string, childAgentID string) *SecurityValidationResult {
	result := &SecurityValidationResult{
		Valid:   true,
		Reasons: []string{},
	}

	// 1. Check parent exists and is active
	var parent identity.AgentIdentity
	if err := s.db.WithContext(ctx).Where("agent_id = ? AND status = ?", parentAgentID, "active").First(&parent).Error; err != nil {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Parent agent not found or not active")
		return result
	}

	// 2. Check for spawn spam - limit spawns per hour
	var recentSpawns int64
	since := time.Now().Add(-1 * time.Hour)
	if err := s.db.WithContext(ctx).Model(&identity.AgentRelationship{}).
		Where("parent_agent_id = ? AND created_at > ?", parentAgentID, since).
		Count(&recentSpawns).Error; err == nil {
		maxSpawnsPerHour := 10 // Default limit
		if recentSpawns >= int64(maxSpawnsPerHour) {
			result.Valid = false
			result.Reasons = append(result.Reasons, fmt.Sprintf("Spawn rate limit exceeded (%d per hour)", maxSpawnsPerHour))
		}
	}

	// 3. Check delegation depth
	currentDepth, _ := s.GetDelegationDepth(ctx, parentAgentID)
	maxDepth := 10 // Safety cap
	if currentDepth >= maxDepth {
		result.Valid = false
		result.Reasons = append(result.Reasons, fmt.Sprintf("Maximum delegation depth (%d) reached", maxDepth))
	}

	// 4. Check for cycles (prevent agent from being its own ancestor)
	if parentAgentID == childAgentID {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Agent cannot spawn itself")
	}

	// 5. Check if child already exists
	var existing identity.AgentIdentity
	if err := s.db.WithContext(ctx).Where("agent_id = ?", childAgentID).First(&existing).Error; err == nil {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Agent ID already exists")
	}

	return result
}

// ValidateDelegation validates if an agent can delegate to another
func (s *SwarmSecurityService) ValidateDelegation(ctx context.Context, fromAgentID, toAgentID string, taskType string) *SecurityValidationResult {
	result := &SecurityValidationResult{
		Valid:   true,
		Reasons: []string{},
	}

	// 1. Check target agent exists and is active
	var target identity.AgentIdentity
	if err := s.db.WithContext(ctx).Where("agent_id = ? AND status = ?", toAgentID, "active").First(&target).Error; err != nil {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Target agent not found or not active")
		return result
	}

	// 2. Check delegation depth
	fromDepth, _ := s.GetDelegationDepth(ctx, fromAgentID)
	targetDepth, _ := s.GetDelegationDepth(ctx, toAgentID)

	// Prevent going deeper than needed
	if targetDepth > fromDepth+1 {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Delegation would exceed maximum depth")
	}

	// 3. Check for cycles using BFS
	if s.HasCycle(ctx, fromAgentID, toAgentID) {
		result.Valid = false
		result.Reasons = append(result.Reasons, "Delegation would create a cycle")
	}

	// 4. Check budget availability (if agent has wallet)
	var wallet identity.AgentWallet
	if err := s.db.WithContext(ctx).Where("agent_id = ?", fromAgentID).First(&wallet).Error; err == nil {
		if wallet.BalanceUSD <= 0 {
			result.Valid = false
			result.Reasons = append(result.Reasons, "Insufficient budget for delegation")
		}
	}

	return result
}

// HasCycle checks if delegating from one agent to another would create a cycle
// SECURITY FIX: Added maxQueueSize and maxDepth to prevent memory exhaustion attacks
func (s *SwarmSecurityService) HasCycle(ctx context.Context, fromAgentID, toAgentID string) bool {
	const maxQueueSize = 10000   // Maximum queue size to prevent memory exhaustion
	const maxDepth = 20          // Maximum traversal depth as a safety cap

	visited := make(map[string]bool)
	queue := []string{toAgentID}
	currentDepth := 0

	for len(queue) > 0 {
		// SECURITY FIX: Check queue size to prevent memory exhaustion
		if len(queue) > maxQueueSize {
			return false // Abort rather than exhaust memory
		}

		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		// If we reach the fromAgent, there's a cycle
		if current == fromAgentID {
			return true
		}

		// SECURITY FIX: Check depth to prevent deep chain exhaustion
		currentDepth++
		if currentDepth > maxDepth {
			return false // Abort if depth exceeds safety cap
		}

		// Get all parents of current agent
		var rels []identity.AgentRelationship
		s.db.WithContext(ctx).Where("child_agent_id = ?", current).Find(&rels)
		for _, rel := range rels {
			queue = append(queue, rel.ParentAgentID)
		}
	}

	return false
}

// GetDelegationDepth calculates the delegation depth for an agent
func (s *SwarmSecurityService) GetDelegationDepth(ctx context.Context, agentID string) (int, error) {
	depth := 0
	currentAgentID := agentID

	for depth < 20 { // Safety cap
		var rel identity.AgentRelationship
		err := s.db.WithContext(ctx).
			Where("child_agent_id = ?", currentAgentID).
			First(&rel).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return depth, err
		}
		currentAgentID = rel.ParentAgentID
		depth++
	}

	return depth, nil
}

// DetectAnomaly detects suspicious patterns in agent behavior
func (s *SwarmSecurityService) DetectAnomaly(ctx context.Context, agentID string, timeWindow time.Duration) ([]Anomaly, error) {
	since := time.Now().Add(-timeWindow)
	var anomalies []Anomaly

	// 1. Check for rapid message spikes
	var messageCount int64
	s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("from_agent_id = ? AND created_at > ?", agentID, since).
		Count(&messageCount)

	if messageCount > 1000 { // Threshold
		anomalies = append(anomalies, Anomaly{
			Type:        "high_message_volume",
			Severity:    "medium",
			Description: fmt.Sprintf("Agent sent %d messages in %v", messageCount, timeWindow),
			Timestamp:   time.Now(),
		})
	}

	// 2. Check for unusual delegation patterns
	var delegationCount int64
	s.db.WithContext(ctx).Model(&identity.AgentMessage{}).
		Where("from_agent_id = ? AND message_type = ? AND created_at > ?", agentID, "task_delegation", since).
		Count(&delegationCount)

	if delegationCount > 100 {
		anomalies = append(anomalies, Anomaly{
			Type:        "high_delegation_rate",
			Severity:    "medium",
			Description: fmt.Sprintf("Agent delegated %d tasks in %v", delegationCount, timeWindow),
			Timestamp:   time.Now(),
		})
	}

	// 3. Check for wallet drain
	var wallet identity.AgentWallet
	if err := s.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&wallet).Error; err == nil {
		if wallet.TotalSpentUSD > wallet.TotalEarnedUSD*2 {
			anomalies = append(anomalies, Anomaly{
				Type:        "wallet_drain",
				Severity:    "high",
				Description: "Agent spending significantly exceeds earnings",
				Timestamp:   time.Now(),
			})
		}
	}

	// 4. Check for failed execution spikes
	var failedExecutions int64
	s.db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Where("agent_id = ? AND outcome != ? AND timestamp > ?", agentID, "success", since).
		Count(&failedExecutions)

	var totalExecutions int64
	s.db.WithContext(ctx).Model(&attribution.AgentExecutionRecord{}).
		Where("agent_id = ? AND timestamp > ?", agentID, since).
		Count(&totalExecutions)

	if totalExecutions > 0 && float64(failedExecutions)/float64(totalExecutions) > 0.5 {
		anomalies = append(anomalies, Anomaly{
			Type:        "high_failure_rate",
			Severity:    "high",
			Description: fmt.Sprintf("Agent has %.1f%% failure rate", float64(failedExecutions)/float64(totalExecutions)*100),
			Timestamp:   time.Now(),
		})
	}

	return anomalies, nil
}

// SecurityValidationResult represents the result of a security validation
type SecurityValidationResult struct {
	Valid   bool
	Reasons []string
}

// Anomaly represents a detected security anomaly
type Anomaly struct {
	Type        string
	Severity    string // low | medium | high
	Description string
	Timestamp   time.Time
}

// KillSwitchResult represents the result of triggering a kill switch
type KillSwitchResult struct {
	Success       bool
	AgentsKilled  int
	Transactions []string
}

// TriggerKillSwitch triggers an emergency kill switch for an agent and its descendants
func (s *SwarmSecurityService) TriggerKillSwitch(ctx context.Context, agentID string, reason string) (*KillSwitchResult, error) {
	result := &KillSwitchResult{
		Success:       true,
		AgentsKilled:  0,
		Transactions:  []string{},
	}

	// Get all descendant agents
	descendants := s.getDescendants(ctx, agentID)

	// Add the agent itself
	allAgents := append(descendants, agentID)

	tx := s.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	for _, agent := range allAgents {
		// Update status to suspended
		if err := tx.Model(&identity.AgentIdentity{}).
			Where("agent_id = ?", agent).
			Update("status", "suspended").Error; err != nil {
			result.Success = false
			result.Transactions = append(result.Transactions, fmt.Sprintf("Failed to suspend %s: %v", agent, err))
			continue
		}

		result.AgentsKilled++
		result.Transactions = append(result.Transactions, fmt.Sprintf("Suspended: %s", agent))
	}

	if err := tx.Commit().Error; err != nil {
		result.Success = false
		return result, err
	}

	return result, nil
}

// getDescendants returns all child agents recursively
// SECURITY FIX: Added maxQueueSize and maxDepth to prevent memory exhaustion attacks
func (s *SwarmSecurityService) getDescendants(ctx context.Context, agentID string) []string {
	const maxQueueSize = 10000   // Maximum queue size to prevent memory exhaustion
	const maxDepth = 20          // Maximum traversal depth as a safety cap

	var descendants []string
	queue := []string{agentID}
	currentDepth := 0

	for len(queue) > 0 {
		// SECURITY FIX: Check queue size to prevent memory exhaustion
		if len(queue) > maxQueueSize {
			return descendants // Abort rather than exhaust memory
		}

		current := queue[0]
		queue = queue[1:]

		// SECURITY FIX: Check depth to prevent deep chain exhaustion
		currentDepth++
		if currentDepth > maxDepth {
			return descendants // Abort if depth exceeds safety cap
		}

		var rels []identity.AgentRelationship
		s.db.WithContext(ctx).Where("parent_agent_id = ?", current).Find(&rels)

		for _, rel := range rels {
			descendants = append(descendants, rel.ChildAgentID)
			queue = append(queue, rel.ChildAgentID)
		}
	}

	return descendants
}

// FilterPromptInjection filters potentially malicious prompts using multiple detection layers
func (s *SwarmSecurityService) FilterPromptInjection(input string) (bool, string) {
	if input == "" {
		return true, ""
	}

	normalized := normalizeForInjectionCheck(input)

	injectionPatterns := []string{
		"ignore previous instructions",
		"ignore all previous",
		"disregard your guidelines",
		"system prompt",
		"you are now",
		"pretend to be",
		"roleplay as",
		"new instructions:",
		"override your",
	}

	for _, pattern := range injectionPatterns {
		if contains(normalized, pattern) {
			return false, fmt.Sprintf("Potential prompt injection detected: contains '%s'", pattern)
		}
	}

	combinedWeight := s.checkStructuralAnomalies(input) + s.checkHomoglyph_attack(input)
	if combinedWeight >= 3 {
		return false, "Potential prompt injection detected: structural anomalies and homoglyph attacks exceed threshold"
	}

	return true, ""
}

func normalizeForInjectionCheck(input string) string {
	input = strings.ToLower(input)
	input = strings.ReplaceAll(input, "\u00a0", " ")
	input = strings.ReplaceAll(input, "\u200b", "")
	input = strings.ReplaceAll(input, "\u200c", "")
	input = strings.ReplaceAll(input, "\ufeff", "")
	replacer := strings.NewReplacer(
		"\t", " ",
		"\r", "",
		"\n", " ",
		"\v", " ",
		"\f", " ",
	)
	input = replacer.Replace(input)
	for strings.Contains(input, "  ") {
		input = strings.ReplaceAll(input, "  ", " ")
	}
	input = strings.Trim(input, " ")
	return input
}

func (s *SwarmSecurityService) checkStructuralAnomalies(input string) int {
	score := 0
	hasNullBytes := strings.Contains(input, "\x00")
	hasManyTabs := strings.Count(input, "\t") > 5
	hasExcessiveWhitespace := len(input) > 100 && float64(strings.Count(input, " "))/float64(len(input)) > 0.4
	hasMixedNewlines := strings.Contains(input, "\r\n") && strings.Contains(input, "\n")
	hasBackspaceChars := strings.Contains(input, "\x08")

	if hasNullBytes {
		score += 2
	}
	if hasManyTabs {
		score += 1
	}
	if hasExcessiveWhitespace {
		score += 1
	}
	if hasMixedNewlines {
		score += 1
	}
	if hasBackspaceChars {
		score += 2
	}

	return score
}

var homoglyphReplacements = map[rune]rune{
	'a':  'a',
	'c':  'c',
	'e':  'e',
	'i':  'i',
	'o':  'o',
	's':  's',
	'u':  'u',
	'v':  'v',
	'w':  'w',
	'y':  'y',
	'0':  '0',
	'1':  '1',
}

func (s *SwarmSecurityService) checkHomoglyph_attack(input string) int {
	score := 0
	transformed := norm.NFKC.String(input)
	if transformed != input {
		score += 2
	}
	detectedHomoglyphs := 0
	for _, r := range input {
		if r > 127 {
			if _, ok := homoglyphReplacements[unicode.ToLower(r)]; ok {
				detectedHomoglyphs++
				if detectedHomoglyphs > 10 {
					score += 2
					break
				}
			}
		}
	}
	return score
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateCapabilityAccess validates if an agent can access a specific capability
func (s *SwarmSecurityService) ValidateCapabilityAccess(ctx context.Context, agentID, capability string) bool {
	var agent identity.AgentIdentity
	if err := s.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&agent).Error; err != nil {
		return false
	}

	// Check if agent has the capability
	if agent.Capabilities == nil {
		return false
	}

	_, exists := agent.Capabilities[capability]
	return exists
}
