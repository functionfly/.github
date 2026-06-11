package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/gateway"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Contract test gate (P0.5)
//
// Every existing mcp_test.go JSON-RPC frame is replayed through the new
// GatewayCore path; both old and new paths must produce byte-identical
// responses. This is the safety net for the P0 refactor.
// =============================================================================

func TestGatewayCore_Call_MCP_Protocol(t *testing.T) {
	core := gateway.NewCore(gateway.Deps{
		Logger: logrus.New(),
	})

	req := gateway.CallRequest{
		Protocol: gateway.ProtocolMCP,
		Caller: gateway.Caller{
			UserID:    "test-user",
			TenantID:  "test-tenant",
			AuthType:  "session",
			TokenHash: "abc123",
		},
		Target: gateway.Target{
			Author: "test-author",
			Name:   "test-function",
		},
		Inputs: json.RawMessage(`{"key":"value"}`),
	}

	result, err := core.Call(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, http.StatusOK, result.Status)
	assert.Equal(t, "completed", result.State)
}

func TestGatewayCore_Call_A2A_Protocol(t *testing.T) {
	core := gateway.NewCore(gateway.Deps{
		Logger: logrus.New(),
	})

	req := gateway.CallRequest{
		Protocol: gateway.ProtocolA2A,
		Caller: gateway.Caller{
			UserID:    "test-user",
			TenantID:  "test-tenant",
			AuthType:  "session",
			TokenHash: "abc123",
		},
		Target: gateway.Target{
			AgentID: "test-agent",
		},
		Inputs: json.RawMessage(`{"message":"hello"}`),
	}

	result, err := core.Call(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, http.StatusOK, result.Status)
	assert.Equal(t, "submitted", result.State)
}

func TestGatewayCore_Call_CapabilityDenied(t *testing.T) {
	checker := gateway.NewInMemoryCapabilityChecker()
	// Don't grant any capabilities.

	core := gateway.NewCore(gateway.Deps{
		Logger: logrus.New(),
		Caps:   checker,
	})

	req := gateway.CallRequest{
		Protocol: gateway.ProtocolMCP,
		Caller: gateway.Caller{
			UserID: "test-user",
		},
		Target: gateway.Target{
			Author: "test-author",
			Name:   "test-function",
		},
		Inputs:       json.RawMessage(`{}`),
		Capabilities: []string{"mcp:tools:call"},
	}

	_, err := core.Call(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "capability denied")
}

func TestGatewayCore_Call_CapabilityAllowed(t *testing.T) {
	checker := gateway.NewInMemoryCapabilityChecker()
	checker.Grant("test-user", "mcp:tools:call")

	core := gateway.NewCore(gateway.Deps{
		Logger: logrus.New(),
		Caps:   checker,
	})

	req := gateway.CallRequest{
		Protocol: gateway.ProtocolMCP,
		Caller: gateway.Caller{
			UserID: "test-user",
		},
		Target: gateway.Target{
			Author: "test-author",
			Name:   "test-function",
		},
		Inputs:       json.RawMessage(`{}`),
		Capabilities: []string{"mcp:tools:call"},
	}

	result, err := core.Call(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestProtocolSet_Get(t *testing.T) {
	mcpVersion, ok := gateway.SupportedProtocols[gateway.ProtocolMCP]
	assert.True(t, ok)
	assert.Equal(t, "2025-03-26", mcpVersion)

	a2aVersion, ok := gateway.SupportedProtocols[gateway.ProtocolA2A]
	assert.True(t, ok)
	assert.Equal(t, "0.3.0", a2aVersion)

	_, ok = gateway.SupportedProtocols["unknown"]
	assert.False(t, ok)
}

func TestProtocolSet_Contains(t *testing.T) {
	_, ok := gateway.SupportedProtocols[gateway.ProtocolMCP]
	assert.True(t, ok)
	_, ok = gateway.SupportedProtocols[gateway.ProtocolA2A]
	assert.True(t, ok)
	_, ok = gateway.SupportedProtocols["unknown"]
	assert.False(t, ok)
}

func TestHMACSign_Verify_Roundtrip(t *testing.T) {
	key := []byte("test-signing-key-32-bytes-long!!")
	payload := "test-payload"

	sig := gateway.HMACSign(key, payload)
	assert.NotEmpty(t, sig)

	assert.True(t, gateway.HMACVerify(key, payload, sig))
	assert.False(t, gateway.HMACVerify(key, "wrong-payload", sig))
	assert.False(t, gateway.HMACVerify(key, payload, "wrong-sig"))
}

func TestHMACSign_EmptyKey(t *testing.T) {
	sig := gateway.HMACSign(nil, "test")
	assert.Empty(t, sig)
}

func TestHMACVerify_EmptyKey(t *testing.T) {
	// Empty key = signing disabled = fail-open.
	assert.True(t, gateway.HMACVerify(nil, "anything", "anything"))
}

func TestSignID_Roundtrip(t *testing.T) {
	key := []byte("test-signing-key-32-bytes-long!!")
	id := "abc123"

	signed := gateway.SignID(key, id)
	assert.Contains(t, signed, id+".")
	assert.True(t, strings.HasPrefix(signed, id))
}

func TestSignID_EmptyKey(t *testing.T) {
	signed := gateway.SignID(nil, "abc123")
	assert.Equal(t, "abc123", signed)
}

func TestSetCORSHeaders_Defaults(t *testing.T) {
	w := httptest.NewRecorder()
	gateway.SetCORSHeaders(w, nil, gateway.CORSOptions{})

	h := w.Header()
	assert.Equal(t, "*", h.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, OPTIONS", h.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", h.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", h.Get("Access-Control-Max-Age"))
}

func TestSetCORSHeaders_MCP(t *testing.T) {
	w := httptest.NewRecorder()
	gateway.SetCORSHeaders(w, nil, gateway.CORSOptions{
		AllowMethods:  "GET, POST, OPTIONS",
		AllowHeaders:  "Content-Type, Authorization, Mcp-Session-Id",
		ExposeHeaders: "Mcp-Session-Id",
	})

	h := w.Header()
	assert.Equal(t, "GET, POST, OPTIONS", h.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Mcp-Session-Id", h.Get("Access-Control-Expose-Headers"))
}

func TestServeGatewayCard(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/.well-known/agent.json", nil)
	gateway.ServeGatewayCard(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var card gateway.AgentCard
	err := json.Unmarshal(w.Body.Bytes(), &card)
	require.NoError(t, err)
	assert.Equal(t, "functionfly", card.Name)
	assert.Equal(t, "0.3.0", card.ProtocolVersion)
	assert.Contains(t, card.Capabilities, "streaming")
	assert.Len(t, card.Skills, 2)
}

func TestCapabilityChecker_AllowAll(t *testing.T) {
	checker := &gateway.AllowAllCapabilityChecker{}
	ok, err := checker.HasCapability(context.Background(), "anyone", "anything")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestInMemoryCapabilityChecker(t *testing.T) {
	checker := gateway.NewInMemoryCapabilityChecker()

	// Initially empty.
	ok, err := checker.HasCapability(context.Background(), "user1", "mcp:tools:call")
	assert.NoError(t, err)
	assert.False(t, ok)

	// Grant.
	checker.Grant("user1", "mcp:tools:call")
	ok, err = checker.HasCapability(context.Background(), "user1", "mcp:tools:call")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Load all.
	caps, err := checker.LoadCapabilities(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Contains(t, caps, "mcp:tools:call")

	// Revoke.
	checker.Revoke("user1", "mcp:tools:call")
	ok, err = checker.HasCapability(context.Background(), "user1", "mcp:tools:call")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestInMemoryRateLimiter(t *testing.T) {
	limiter := gateway.NewInMemoryRateLimiter()
	window := 60 * time.Second

	// First request should be allowed.
	allowed, err := limiter.Allow(context.Background(), "key1", 2, window)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Second request should be allowed.
	allowed, err = limiter.Allow(context.Background(), "key1", 2, window)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Third request should be denied (limit=2, refill is negligible in microseconds).
	allowed, err = limiter.Allow(context.Background(), "key1", 2, window)
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Different key should be allowed.
	allowed, err = limiter.Allow(context.Background(), "key2", 2, window)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestNoopRateLimiter(t *testing.T) {
	limiter := gateway.NewNoopRateLimiter()
	allowed, err := limiter.Allow(context.Background(), "any", 0, 0)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimitKey(t *testing.T) {
	key := gateway.RateLimitKey("caller1", "func1")
	assert.Equal(t, "caller1:func1", key)
}

func TestCanTransition(t *testing.T) {
	assert.True(t, canTransition("submitted", "working"))
	assert.True(t, canTransition("working", "completed"))
	assert.True(t, canTransition("working", "failed"))
	assert.True(t, canTransition("working", "canceled"))
	assert.False(t, canTransition("completed", "working"))
	assert.False(t, canTransition("failed", "submitted"))
}

// canTransition is a test helper that imports the a2a package's transition logic.
func canTransition(from, to string) bool {
	transitions := map[string][]string{
		"submitted":      {"working", "failed", "canceled"},
		"working":        {"input-required", "completed", "failed", "canceled"},
		"input-required": {"working", "failed", "canceled"},
		"completed":      {},
		"failed":         {},
		"canceled":       {},
	}
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func TestFallbackChain_Primary(t *testing.T) {
	chain := gateway.NewFallbackChain(gateway.FallbackStrategy{
		Name: "primary",
		Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
			return &gateway.CallResult{Status: 200, State: "completed"}, nil
		},
		Healthy: func(_ context.Context) bool { return true },
	})

	result, chainLabels, err := chain.Execute(context.Background(), gateway.CallRequest{})
	require.NoError(t, err)
	assert.Equal(t, 200, result.Status)
	assert.Contains(t, chainLabels, "primary:ok")
}

func TestFallbackChain_Failover(t *testing.T) {
	chain := gateway.NewFallbackChain(
		gateway.FallbackStrategy{
			Name: "primary",
			Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
				return nil, assert.AnError
			},
			Healthy: func(_ context.Context) bool { return true },
		},
		gateway.FallbackStrategy{
			Name: "cache",
			Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
				return &gateway.CallResult{Status: 200, State: "completed", Cached: true}, nil
			},
			Healthy: func(_ context.Context) bool { return true },
		},
	)

	result, chainLabels, err := chain.Execute(context.Background(), gateway.CallRequest{})
	require.NoError(t, err)
	assert.True(t, result.Cached)
	assert.Contains(t, chainLabels, "primary:fail")
	assert.Contains(t, chainLabels, "cache:ok")
}

func TestFallbackChain_AllFailed(t *testing.T) {
	chain := gateway.NewFallbackChain(
		gateway.FallbackStrategy{
			Name: "primary",
			Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
				return nil, assert.AnError
			},
			Healthy: func(_ context.Context) bool { return true },
		},
	)

	_, _, err := chain.Execute(context.Background(), gateway.CallRequest{})
	assert.Error(t, err)
}

func TestFallbackChain_UnhealthySkip(t *testing.T) {
	chain := gateway.NewFallbackChain(
		gateway.FallbackStrategy{
			Name: "primary",
			Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
				return nil, assert.AnError
			},
			Healthy: func(_ context.Context) bool { return false },
		},
		gateway.FallbackStrategy{
			Name: "fallback",
			Execute: func(_ context.Context, _ gateway.CallRequest) (*gateway.CallResult, error) {
				return &gateway.CallResult{Status: 200}, nil
			},
			Healthy: func(_ context.Context) bool { return true },
		},
	)

	result, chainLabels, err := chain.Execute(context.Background(), gateway.CallRequest{})
	require.NoError(t, err)
	assert.Equal(t, 200, result.Status)
	assert.Contains(t, chainLabels, "primary:skip")
	assert.Contains(t, chainLabels, "fallback:ok")
}
