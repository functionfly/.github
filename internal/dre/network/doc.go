// Package network implements network mode handling for the DCC Protocol.
//
// The NetworkHandler provides three modes for deterministic network behavior:
//
// Record Mode:
//   - First execution: allow outbound, record request/response
//   - Replay: return recorded response
//   - Store response hash in DependencyHash
//
// Stub Mode:
//   - Load from stub_manifest.json
//   - Map URL → deterministic response payload
//   - No network calls made
//
// Disabled Mode:
//   - Abort all network calls
//   - Maximum determinism, no external dependencies
//
// Usage:
//
//	handler := network.New(network.ModeRecord)
//	
//	// In record mode, record responses
//	handler.Record("https://api.example.com/data", []byte(`{"result": "ok"}`))
//	
//	// Handle requests
//	response, err := handler.HandleRequest("https://api.example.com/data", nil)
//	
//	// Get dependency hash for MEG
//	depHash := handler.DependencyHash()
package network
