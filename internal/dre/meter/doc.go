// Package meter implements instruction metering for the DCC Protocol.
//
// The Meter tracks resource usage during capsule execution including:
//   - Total instructions executed
//   - Peak memory usage
//   - Syscall count
//   - Output → ResourceHash
//
// Usage:
//
//	limits := &meter.Limits{
//	    MaxInstructions: 1_000_000_000,
//	    MaxMemory:       128 * 1024 * 1024,
//	    MaxSyscalls:     1_000_000,
//	}
//	
//	m := meter.New(limits)
//	
//	// Track execution
//	m.IncInstructions()
//	m.SetMemory(currentMemoryBytes)
//	m.IncSyscallCount()
//	
//	// Check limits
//	if !m.IsWithinLimits() {
//	    // Handle limit exceeded
//	}
//	
//	// Get resource hash for MEG
//	resourceHash := m.ResourceHash()
package meter
