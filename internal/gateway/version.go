package gateway

// SupportedProtocols declares the protocol versions this gateway speaks.
// Pin versions here — the well-known manifests, Agent Cards, and MCP
// ServerManifest all source from this single location.
var SupportedProtocols = map[Protocol]string{
	ProtocolMCP: "2025-03-26",
	ProtocolA2A: "0.3.0",
}
