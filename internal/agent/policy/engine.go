package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// BehavioralPolicy defines the safety constraints for an agent
type BehavioralPolicy struct {
	AgentID             string   `json:"agent_id" gorm:"uniqueIndex;not null"`
	MaxExecutionDepth   int      `json:"max_execution_depth" gorm:"not null;default:10"`
	MaxRecursionDepth   int      `json:"max_recursion_depth" gorm:"not null;default:3"`
	MaxWallTimeMs       int      `json:"max_wall_time_ms" gorm:"not null;default:300000"`
	MaxMemoryGrowthMB   int      `json:"max_memory_growth_mb" gorm:"not null;default:512"`
	ForbiddenFunctions  []string `json:"forbidden_functions,omitempty" gorm:"type:text[]"`
	DeterministicOnly   bool     `json:"deterministic_only" gorm:"not null;default:false"`
	AllowedCapabilities []string `json:"allowed_capabilities,omitempty" gorm:"type:text[]"`
}

// TableName returns the GORM table name
func (BehavioralPolicy) TableName() string {
	return "agent_behavioral_policies"
}

// PolicyResult is the result of a policy check
type PolicyResult struct {
	Allowed   bool              `json:"allowed"`
	Violation *PolicyViolation  `json:"violation,omitempty"`
}

// PolicyViolation describes a policy violation
type PolicyViolation struct {
	Code    PolicyViolationCode `json:"code"`
	Message string              `json:"message"`
}

// PolicyViolationCode represents the type of policy violation
type PolicyViolationCode string

const (
	ViolationLoopDetected    PolicyViolationCode = "LOOP_DETECTED"
	ViolationDepthExceeded   PolicyViolationCode = "DEPTH_EXCEEDED"
	ViolationFunctionBlocked PolicyViolationCode = "FUNCTION_BLOCKED"
	ViolationCapabilityDenied PolicyViolationCode = "CAPABILITY_DENIED"
	ViolationDeterministicOnly PolicyViolationCode = "DETERMINISTIC_ONLY"
)

func (v *PolicyViolation) Error() string {
	return string(v.Code) + ": " + v.Message
}

// AgentExecutionRequest is the input to a policy check
type AgentExecutionRequest struct {
	AgentID     string          `json:"agent_id"`
	SessionID   string          `json:"session_id"`
	FunctionURI string          `json:"function_uri"`
	Input       json.RawMessage `json:"input"`
	CallDepth   int             `json:"call_depth"`
	Capabilities []string       `json:"capabilities,omitempty"`
	Deterministic bool          `json:"deterministic"`
}

// Engine evaluates behavioral policies for agent executions
type Engine struct {
	redis *redis.Client
	db    *gorm.DB
}

// NewEngine creates a new policy engine
func NewEngine(db *gorm.DB, redisClient *redis.Client) *Engine {
	return &Engine{
		redis: redisClient,
		db:    db,
	}
}

// CheckPolicy evaluates all behavioral policies for an agent execution request
func (e *Engine) CheckPolicy(ctx context.Context, agentID string, req *AgentExecutionRequest) (*PolicyResult, error) {
	// Load policy from DB
	policy, err := e.loadPolicy(ctx, agentID)
	if err != nil {
		// If no policy configured, use permissive defaults
		policy = defaultPolicy(agentID)
	}

	// 1. Check forbidden functions
	for _, forbidden := range policy.ForbiddenFunctions {
		if matchPattern(req.FunctionURI, forbidden) {
			return &PolicyResult{
				Allowed: false,
				Violation: &PolicyViolation{
					Code:    ViolationFunctionBlocked,
					Message: fmt.Sprintf("function %s is blocked by agent policy", req.FunctionURI),
				},
			}, nil
		}
	}

	// 2. Check deterministic-only policy
	if policy.DeterministicOnly && !req.Deterministic {
		return &PolicyResult{
			Allowed: false,
			Violation: &PolicyViolation{
				Code:    ViolationDeterministicOnly,
				Message: fmt.Sprintf("agent policy requires deterministic functions only; %s is non-deterministic", req.FunctionURI),
			},
		}, nil
	}

	// 3. Check allowed capabilities
	if len(policy.AllowedCapabilities) > 0 {
		for _, cap := range req.Capabilities {
			if !containsString(policy.AllowedCapabilities, cap) {
				return &PolicyResult{
					Allowed: false,
					Violation: &PolicyViolation{
						Code:    ViolationCapabilityDenied,
						Message: fmt.Sprintf("capability %s is not allowed by agent policy", cap),
					},
				}, nil
			}
		}
	}

	// 4. Check execution depth
	if req.CallDepth >= policy.MaxExecutionDepth {
		return &PolicyResult{
			Allowed: false,
			Violation: &PolicyViolation{
				Code:    ViolationDepthExceeded,
				Message: fmt.Sprintf("execution depth %d exceeds maximum %d", req.CallDepth, policy.MaxExecutionDepth),
			},
		}, nil
	}

	// 5. Check for loops (if Redis is available)
	if e.redis != nil && req.SessionID != "" {
		loopDetected, err := e.DetectLoop(ctx, agentID, req.SessionID, req.FunctionURI, req.Input)
		if err == nil && loopDetected {
			return &PolicyResult{
				Allowed: false,
				Violation: &PolicyViolation{
					Code:    ViolationLoopDetected,
					Message: fmt.Sprintf("loop detected: function %s called with identical input more than %d times in this session", req.FunctionURI, loopThreshold),
				},
			}, nil
		}
	}

	return &PolicyResult{Allowed: true}, nil
}

const loopThreshold = 3

// DetectLoop checks if a (functionURI, inputHash) pair has been called too many times in a session
func (e *Engine) DetectLoop(ctx context.Context, agentID, sessionID, functionURI string, input json.RawMessage) (bool, error) {
	if e.redis == nil {
		return false, nil
	}

	inputHash := hashInput(input)
	key := fmt.Sprintf("policy:loop:%s:%s:%s:%s", agentID, sessionID, functionURI, inputHash)

	count, err := e.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set TTL on first call (session duration = 1 hour)
	if count == 1 {
		e.redis.Expire(ctx, key, time.Hour)
	}

	return count > int64(loopThreshold), nil
}

// TrackDepth increments and returns the current call depth for a session
func (e *Engine) TrackDepth(ctx context.Context, agentID, sessionID string) (int, error) {
	if e.redis == nil {
		return 0, nil
	}

	key := fmt.Sprintf("policy:depth:%s:%s", agentID, sessionID)
	count, err := e.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		e.redis.Expire(ctx, key, time.Hour)
	}

	return int(count), nil
}

// DecrDepth decrements the call depth when an execution completes
func (e *Engine) DecrDepth(ctx context.Context, agentID, sessionID string) {
	if e.redis == nil {
		return
	}
	key := fmt.Sprintf("policy:depth:%s:%s", agentID, sessionID)
	e.redis.Decr(ctx, key)
}

// UpsertPolicy creates or updates a behavioral policy for an agent
func (e *Engine) UpsertPolicy(ctx context.Context, policy *BehavioralPolicy) error {
	return e.db.WithContext(ctx).
		Where("agent_id = ?", policy.AgentID).
		Assign(policy).
		FirstOrCreate(policy).Error
}

// GetPolicy retrieves the behavioral policy for an agent
func (e *Engine) GetPolicy(ctx context.Context, agentID string) (*BehavioralPolicy, error) {
	return e.loadPolicy(ctx, agentID)
}

func (e *Engine) loadPolicy(ctx context.Context, agentID string) (*BehavioralPolicy, error) {
	var policy BehavioralPolicy
	err := e.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func defaultPolicy(agentID string) *BehavioralPolicy {
	return &BehavioralPolicy{
		AgentID:           agentID,
		MaxExecutionDepth: 10,
		MaxRecursionDepth: 3,
		MaxWallTimeMs:     300000,
		MaxMemoryGrowthMB: 512,
		DeterministicOnly: false,
	}
}

func hashInput(input json.RawMessage) string {
	h := sha256.Sum256(input)
	return hex.EncodeToString(h[:8]) // Use first 8 bytes for key brevity
}

func matchPattern(uri, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(uri, prefix+"/")
	}
	return uri == pattern
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
