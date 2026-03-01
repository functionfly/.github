// Package trace implements trace recording for the DCC Protocol.
//
// The Trace records execution traces as chunks that are included in TraceHash
// for deterministic verification.
//
// Key features:
//   - Records each syscall as: H("FX_SYSCALL" || syscall_type || canonical_args || result)
//   - Chunks are linked via previous hash (Merkle-like structure)
//   - Included in TraceHash
//   - Lite tier: disable detailed tracing
//
// Usage:
//
//	tr := trace.New(1000) // 1000 entries per chunk
//	
//	// Record execution events
//	tr.Record("syscall", virtualTime, []byte(`{"type": "memory_alloc"}`))
//	tr.Record("memory", virtualTime, []byte(`{"addr": "0x1000"}`))
//	
//	// Get trace hash for MEG
//	traceHash := tr.Hash()
//	
//	// For lite tier, disable detailed tracing
//	tr.Disable()
package trace
