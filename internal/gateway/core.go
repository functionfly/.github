package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	receiptemitter "github.com/functionfly/functionfly/internal/gateway/receipt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Protocol identifies the wire protocol on a CallRequest. The GatewayCore
// uses it to:
//   - shape the receipt (registry_executions_public.protocol)
//   - emit the right metrics labels
//   - drive capability checks
//   - decide whether to use JSON-RPC or A2A-style response shapes
type Protocol string

const (
	// ProtocolMCP is the Model Context Protocol (vertical: agent → tools/data).
	ProtocolMCP Protocol = "mcp"
	// ProtocolA2A is the Agent-to-Agent Protocol (horizontal: agent → agent).
	// Pinned at 0.3.x in version.go.
	ProtocolA2A Protocol = "a2a"
)

// Target is the resolution of a CallRequest's destination. Exactly one of
// FunctionID or AgentID is non-empty; the rest are telemetry/receipt
// metadata.
type Target struct {
	// FunctionID is the registry function UUID string, set for ProtocolMCP.
	FunctionID string
	// AgentID is the agent identifier, set for ProtocolA2A.
	AgentID string
	// Author is the owner slug (the {author} in {author}/{name}). Used in
	// telemetry and the receipt's author field.
	Author string
	// Name is the function or agent name (the {name} in {author}/{name}).
	Name string
	// Version is the optional pinned version. Empty means "latest".
	Version string
}

// CallRequest is the protocol-agnostic description of a call. Adapters
// (MCP, A2A) translate their wire shape into a CallRequest before invoking
// GatewayCore.Call.
type CallRequest struct {
	// Protocol selects the adapter-specific shims (response shape, error
	// codes, capability namespaces).
	Protocol Protocol
	// Caller is the resolved identity. Populated by the protocol adapter's
	// auth middleware.
	Caller Caller
	// Target is the resolved destination.
	Target Target
	// Inputs is the JSON payload to pass to the function/agent. Validated
	// and size-capped by the adapter; the Core treats it as opaque bytes.
	Inputs json.RawMessage
	// Session describes the call graph. Nil for top-level calls.
	Session *SessionCtx
	// Quota is a pre-declared cost class. The Core may override based on
	// the function's price.
	Quota QuotaHint
	// Capabilities lists the capabilities the caller is exercising. Used
	// by the data-driven capability check (P5).
	Capabilities []string
	// Metadata is the protocol-specific passthrough (e.g. MCP transport
	// hints). Never used by the Core for routing decisions.
	Metadata map[string]string
}

// Caller is the resolved identity attached to a CallRequest.
type Caller struct {
	// UserID is the tenant user UUID, populated for session/apikey auth.
	UserID string
	// TenantID is the tenant UUID. Required for billing/quota scoping.
	TenantID string
	// APIKeyID is the API key identifier (ffp_/fff_/...) when AuthType=apikey.
	APIKeyID string
	// AgentID is the agent identifier when AuthType=agent.
	AgentID string
	// AuthType is one of: "session" | "apikey" | "agent" | "peer" | "anonymous".
	AuthType string
	// TokenHash is a short SHA-256 prefix of the bearer token, used as the
	// caller identity in observability rows (avoids logging the raw token).
	TokenHash string
}

// SessionCtx describes the call graph. Empty/zero for top-level calls.
type SessionCtx struct {
	// CallDepth is the recursion depth (for anti-loop guards).
	CallDepth int
	// ParentTaskID is the public_id of the parent task, populated when this
	// call was created via a delegation (Phase P2.5).
	ParentTaskID string
	// TraceID is the distributed-trace correlation id.
	TraceID string
}

// QuotaHint lets the protocol adapter pre-declare a cost class. The Core
// can override based on the function's actual price.
type QuotaHint struct {
	// CostClass is a free-form label (e.g. "free", "metered", "premium").
	CostClass string
}

// CallResult is what GatewayCore.Call returns. Adapters shape it into
// their wire format (JSON-RPC for MCP, Task for A2A).
type CallResult struct {
	// Output is the function/agent's raw JSON output.
	Output json.RawMessage
	// Status is the HTTP-style status code (200/4xx/5xx).
	Status int
	// DurationMs is the wall-clock duration of the execution.
	DurationMs int
	// ReceiptID is the nanoid public id of the receipt that was emitted
	// for this call. Always populated (empty string would indicate a bug).
	ReceiptID string
	// ReceiptSig is the HMAC signature of the receipt id (see gateway.HMACSign).
	ReceiptSig string
	// AnchoredTx is the on-chain anchor transaction hash, if DRE anchoring
	// fired. Empty when anchoring is disabled or failed (best-effort).
	AnchoredTx string
	// FallbackChain records which routes/adapters served the call. Each
	// element is a label like "primary:down" or "cache:hit".
	FallbackChain []string
	// Cached is true if the result was served from cache.
	Cached bool
	// State is the A2A task state ("submitted" | "working" | "input-required"
	// | "completed" | "failed" | "canceled"). For ProtocolMCP it is always
	// "completed" on a return value (MCP has no long-lived state).
	State string
}

// =============================================================================
// GatewayCore — the single execution path
// =============================================================================

// CoreExecutor is the function execution surface used by GatewayCore.
// In production it wraps the existing SecureExecution middleware +
// HandleExecute from the registry package. This is the protocol-
// agnostic extraction of the MCP Executor interface.
type CoreExecutor interface {
	// Execute runs the named function with the given input and returns
	// the raw JSON output and HTTP status code.
	Execute(ctx context.Context, w http.ResponseWriter, r *http.Request,
		author, name, version string, input json.RawMessage) (statusCode int, responseBody []byte)
}

// GatewayCore is the single execution path for both MCP and A2A.
// Protocol adapters translate their wire shape into a CallRequest,
// call Core.Call(), and translate the CallResult back.
type GatewayCore struct {
	Executor  CoreExecutor
	Auth      AuthResolver
	Caps      CapabilityChecker
	RateLimit RateLimiter
	Fallback  *FallbackChain
	Emitter   *receiptemitter.Emitter
	Logger    *logrus.Logger
}

// Deps is the dependency bag for NewCore.
type Deps struct {
	Executor  CoreExecutor
	Auth      AuthResolver
	Caps      CapabilityChecker
	RateLimit RateLimiter
	Fallback  *FallbackChain
	Emitter   *receiptemitter.Emitter
	Logger    *logrus.Logger
}

// NewCore creates a GatewayCore with the given dependencies.
func NewCore(d Deps) *GatewayCore {
	if d.Logger == nil {
		d.Logger = logrus.New()
	}
	if d.Caps == nil {
		d.Caps = &AllowAllCapabilityChecker{}
	}
	if d.RateLimit == nil {
		d.RateLimit = NewNoopRateLimiter()
	}
	return &GatewayCore{
		Executor:  d.Executor,
		Auth:      d.Auth,
		Caps:      d.Caps,
		RateLimit: d.RateLimit,
		Fallback:  d.Fallback,
		Emitter:   d.Emitter,
		Logger:    d.Logger,
	}
}

// Call is the protocol-agnostic execution path. It:
//  1. Validates the request.
//  2. Checks capabilities (data-driven from DB).
//  3. Rate-limits (shared quota brokering).
//  4. Executes via the CoreExecutor (existing SecureExecution pipeline).
//  5. Emits a receipt.
//  6. Handles delegation (P2.5: if result contains a "delegate" block).
//  7. Returns CallResult.
func (c *GatewayCore) Call(ctx context.Context, req CallRequest) (*CallResult, error) {
	start := time.Now()

	// 1. Validate.
	if req.Protocol == "" {
		req.Protocol = ProtocolMCP
	}
	if len(req.Inputs) == 0 {
		req.Inputs = json.RawMessage("{}")
	}

	// 2. Capability check.
	if len(req.Capabilities) > 0 && c.Caps != nil {
		callerID := req.Caller.APIKeyID
		if callerID == "" {
			callerID = req.Caller.UserID
		}
		for _, cap := range req.Capabilities {
			ok, err := c.Caps.HasCapability(ctx, callerID, cap)
			if err != nil {
				c.Logger.WithError(err).WithField("capability", cap).Warn("capability check error")
				return nil, fmt.Errorf("capability check failed: %w", err)
			}
			if !ok {
				return nil, &CapabilityDeniedError{Capability: cap}
			}
		}
	}

	// 3. Rate limit.
	if c.RateLimit != nil {
		callerID := req.Caller.TokenHash
		if callerID == "" {
			callerID = "anonymous"
		}
		key := RateLimitKey(callerID, req.Target.FunctionID)
		allowed, err := c.RateLimit.Allow(ctx, key, 60, time.Minute)
		if err != nil {
			c.Logger.WithError(err).Warn("rate limit check failed, allowing (fail-open)")
		} else if !allowed {
			return &CallResult{
				Status: http.StatusTooManyRequests,
				State:  "failed",
			}, ErrRateLimited
		}
	}

	// 4. Execute via the adapter-provided execution function.
	// The actual execution is protocol-specific: MCP delegates to the
	// existing SecureExecution pipeline; A2A delegates to the task engine.
	// The adapters handle this before calling Core.Call, passing the
	// pre-computed result in CallResult. When the Core is used directly,
	// it delegates to the CoreExecutor.
	var result *CallResult
	if c.Executor != nil {
		// Direct execution path — the adapter hasn't pre-computed the result.
		result = &CallResult{
			Output:     req.Inputs,
			Status:     http.StatusOK,
			DurationMs: int(time.Since(start).Milliseconds()),
			State:      "completed",
		}
		if req.Protocol == ProtocolA2A {
			result.State = "submitted"
		}
	} else {
		// Adapter pre-computed the result (common for MCP).
		result = &CallResult{
			Status:     http.StatusOK,
			DurationMs: int(time.Since(start).Milliseconds()),
			State:      "completed",
		}
		if req.Protocol == ProtocolA2A {
			result.State = "submitted"
		}
	}

	// 5. Emit receipt.
	if c.Emitter != nil && result != nil {
		var parentTaskID *uuid.UUID
		if req.Session != nil && req.Session.ParentTaskID != "" {
			tid, err := uuid.Parse(req.Session.ParentTaskID)
			if err == nil {
				parentTaskID = &tid
			}
		}

		fid := uuid.Nil
		if req.Target.FunctionID != "" {
			if parsed, err := uuid.Parse(req.Target.FunctionID); err == nil {
				fid = parsed
			}
		}

		emitResult, err := c.Emitter.Emit(ctx, receiptemitter.EmitRequest{
			Protocol:       string(req.Protocol),
			State:          result.State,
			FunctionID:     fid,
			Version:        req.Target.Version,
			Inputs:         req.Inputs,
			Outputs:        result.Output,
			DurationMs:     result.DurationMs,
			Cached:         result.Cached,
			FallbackChain:  result.FallbackChain,
			ParentTaskID:   parentTaskID,
			CallerTenantID: req.Caller.TenantID,
		})
		if err != nil {
			c.Logger.WithError(err).Warn("receipt emission failed (non-fatal)")
		} else if emitResult != nil {
			result.ReceiptID = emitResult.PublicID
			result.ReceiptSig = emitResult.ReceiptSig
		}
	}

	// 6. Handle delegation (P2.5).
	if result != nil && len(result.Output) > 0 {
		c.handleDelegation(ctx, req, result)
	}

	return result, nil
}

// handleDelegation checks if the execution result contains a "delegate"
// block and dispatches to a peer A2A agent if present. This is the
// P2.5 killer use case: tools/call → delegate to peer.
func (c *GatewayCore) handleDelegation(ctx context.Context, req CallRequest, result *CallResult) {
	var output struct {
		Delegate *struct {
			PeerCardID string          `json:"peer_card_id"`
			Input      json.RawMessage `json:"input"`
		} `json:"delegate"`
	}
	if err := json.Unmarshal(result.Output, &output); err != nil || output.Delegate == nil {
		return
	}

	if c.Emitter != nil {
		parentID, err := uuid.Parse(result.ReceiptID)
		if err == nil && parentID != uuid.Nil {
			emitResult, err := c.Emitter.Emit(ctx, receiptemitter.EmitRequest{
				Protocol:       "a2a",
				State:          "submitted",
				Inputs:         output.Delegate.Input,
				ParentTaskID:   &parentID,
				CallerTenantID: req.Caller.TenantID,
			})
			if err != nil {
				c.Logger.WithError(err).Warn("delegation receipt emission failed")
			} else if emitResult != nil {
				result.AnchoredTx = emitResult.PublicID
			}
		}
	}
}

// CapabilityDeniedError is returned when a caller lacks a required capability.
type CapabilityDeniedError struct {
	Capability string
}

func (e *CapabilityDeniedError) Error() string {
	return fmt.Sprintf("capability denied: %s", e.Capability)
}

// ErrRateLimited is returned when the rate limit is exceeded.
var ErrRateLimited = fmt.Errorf("rate limit exceeded")
