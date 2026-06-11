package gateway

import (
	"encoding/json"
	"net/http"
)

// WellKnownManifest is the extended /.well-known/functionfly.json response
// that includes the supported_protocols block for feature detection.
type WellKnownManifest struct {
	SchemaVersion      string              `json:"schema_version"`
	Provider           string              `json:"provider"`
	SupportedProtocols map[Protocol]string `json:"supported_protocols"`
	Endpoints          WellKnownEndpoints  `json:"endpoints"`
}

// WellKnownEndpoints lists the protocol-specific endpoints.
type WellKnownEndpoints struct {
	MCPStreamableHTTP string `json:"mcp_streamable_http,omitempty"`
	A2ATasksSend      string `json:"a2a_tasks_send,omitempty"`
	A2AAgentCard      string `json:"a2a_agent_card,omitempty"`
	WellKnownAgent    string `json:"well_known_agent,omitempty"`
	ReceiptsPublic    string `json:"receipts_public,omitempty"`
}

// AgentCard is the A2A Agent Card shape served at
// GET /.well-known/agent.json and GET /v1/a2a/agents/{agent_id}/card.
type AgentCard struct {
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	URL                string           `json:"url"`
	Version            string           `json:"version"`
	ProtocolVersion    string           `json:"protocolVersion"`
	Capabilities       []string         `json:"capabilities"`
	Skills             []AgentSkill     `json:"skills"`
	Authentication     AgentAuth        `json:"authentication"`
	DefaultInputModes  []string         `json:"defaultInputModes"`
	DefaultOutputModes []string         `json:"defaultOutputModes"`
}

// AgentSkill is a single skill advertised by an agent.
type AgentSkill struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// AgentAuth describes the authentication schemes an agent supports.
type AgentAuth struct {
	Schemes []string `json:"schemes"`
}

// ServeGatewayCard serves GET /.well-known/agent.json — the FunctionFly
// gateway's own Agent Card.
func ServeGatewayCard(w http.ResponseWriter, r *http.Request) {
	a2aVersion := SupportedProtocols[ProtocolA2A]

	card := AgentCard{
		Name:            "functionfly",
		Description:     "FunctionFly agent gateway — authenticate once, route once, receipt once for both MCP and A2A.",
		URL:             getAPIBase(r) + "/v1/a2a",
		Version:         "1.0",
		ProtocolVersion: a2aVersion,
		Capabilities:    []string{"streaming", "push-notifications"},
		Skills: []AgentSkill{
			{ID: "execute", Description: "Run any registered function"},
			{ID: "delegate", Description: "Hand off a tool result to a peer agent"},
		},
		Authentication: AgentAuth{
			Schemes: []string{"bearer", "agent-apikey"},
		},
		DefaultInputModes:  []string{"application/json"},
		DefaultOutputModes: []string{"application/json"},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	SetCORSHeaders(w, r, CORSOptions{
		AllowMethods: "GET, OPTIONS",
	})
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(card)
}

// ServeWellKnownExtended serves GET /.well-known/functionfly.json with the
// extended supported_protocols block. This is the superset manifest that
// callers use for feature detection.
func ServeWellKnownExtended(w http.ResponseWriter, r *http.Request, base string) {
	if base == "" {
		base = getAPIBase(r)
	}

	manifest := WellKnownManifest{
		SchemaVersion:      "1.0",
		Provider:           "functionfly",
		SupportedProtocols: SupportedProtocols,
		Endpoints: WellKnownEndpoints{
			MCPStreamableHTTP: base + "/v1/mcp",
			A2ATasksSend:      base + "/v1/a2a/{agent_id}/tasks/send",
			A2AAgentCard:      base + "/v1/a2a/agents/{agent_id}/card",
			WellKnownAgent:    base + "/.well-known/agent.json",
			ReceiptsPublic:    base + "/v1/receipts/{id}",
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	SetCORSHeaders(w, r, CORSOptions{
		AllowMethods: "GET, OPTIONS",
	})
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(manifest)
}

// getAPIBase derives the API base URL from the request.
func getAPIBase(r *http.Request) string {
	scheme := "https"
	if r != nil && r.TLS == nil {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto == "http" {
			scheme = "http"
		} else if host := r.Host; host != "" &&
			(len(host) >= 9 && host[:9] == "localhost" ||
				len(host) >= 9 && host[:9] == "127.0.0.1") {
			scheme = "http"
		}
	}
	host := ""
	if r != nil {
		host = r.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = r.Host
		}
	}
	if host == "" {
		host = "api.functionfly.com"
	}
	return scheme + "://" + host
}
