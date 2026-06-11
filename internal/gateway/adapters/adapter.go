// Package adapters provides thin protocol-specific shims on top of
// GatewayCore. Each adapter translates its wire shape (JSON-RPC for
// MCP, REST/JSON for A2A) into a CallRequest and back.
package adapters

import (
	"context"
	"encoding/json"

	"github.com/functionfly/functionfly/internal/gateway"
)

// MCPAdapter is a thin shim that translates MCP JSON-RPC tool calls
// into GatewayCore CallRequests. It is used by the MCP handler to
// delegate execution to the unified gateway path.
type MCPAdapter struct {
	core *gateway.GatewayCore
}

// NewMCPAdapter creates a new MCP adapter.
func NewMCPAdapter(core *gateway.GatewayCore) *MCPAdapter {
	return &MCPAdapter{core: core}
}

// ExecuteToolCall translates an MCP tools/call into a GatewayCore.Call.
// The caller identity and tool resolution are already done by the MCP
// handler; this adapter only handles the execution + receipt emission.
func (a *MCPAdapter) ExecuteToolCall(
	ctx context.Context,
	caller gateway.Caller,
	author, name, version string,
	input json.RawMessage,
	metadata map[string]string,
) (*gateway.CallResult, error) {
	req := gateway.CallRequest{
		Protocol: gateway.ProtocolMCP,
		Caller:   caller,
		Target: gateway.Target{
			Author:  author,
			Name:    name,
			Version: version,
		},
		Inputs:   input,
		Metadata: metadata,
	}

	return a.core.Call(ctx, req)
}

// A2AAdapter is a thin shim that translates A2A task requests into
// GatewayCore CallRequests. Used by the A2A handler.
type A2AAdapter struct {
	core *gateway.GatewayCore
}

// NewA2AAdapter creates a new A2A adapter.
func NewA2AAdapter(core *gateway.GatewayCore) *A2AAdapter {
	return &A2AAdapter{core: core}
}

// ExecuteTask translates an A2A tasks/send into a GatewayCore.Call.
func (a *A2AAdapter) ExecuteTask(
	ctx context.Context,
	caller gateway.Caller,
	agentID string,
	input json.RawMessage,
	metadata map[string]string,
) (*gateway.CallResult, error) {
	req := gateway.CallRequest{
		Protocol: gateway.ProtocolA2A,
		Caller:   caller,
		Target: gateway.Target{
			AgentID: agentID,
		},
		Inputs:   input,
		Metadata: metadata,
	}

	return a.core.Call(ctx, req)
}
