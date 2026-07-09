//go:build cgo

// security.go preserves the legacy file path while delegating the scanner
// implementation to scanner.go (no build tag), so the wazero-backed
// GoRuntime can also use it.
//
// ThreatSeverity, RuntimeThreat, RuntimeScanResult, RuntimeScanner,
// NewRuntimeScanner, ScanSource, ScanBytes, LogThreats, indexOf, and
// hasAnySubstring are defined in scanner.go.
package wasm
