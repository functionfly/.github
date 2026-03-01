// Package syscall implements syscall virtualization for the DCC Protocol.
//
// The SyscallGate (FX_SYSCALL_GATE) controls which system calls are allowed
// in deterministic mode and records syscalls for trace hashing.
//
// Allowed syscalls (deterministic):
//   - Memory allocation
//   - Virtual FS read/write
//   - Virtual clock read
//   - Deterministic RNG
//   - Structured logging
//   - Environment/argument getters
//
// Blocked syscalls (non-deterministic):
//   - Raw socket creation
//   - Direct system clock
//   - Direct disk access
//   - Process spawning
//   - Host file system access
//   - OS random source
//   - Wall time
//   - Process/User ID
//
// Usage:
//
//	gate := syscall.New("strict-v1")
//	
//	// Check if syscall is allowed
//	if gate.Allowed(syscall.SyscallMemoryAlloc) {
//	    // Handle memory allocation
//	}
//	
//	// Record syscall for trace
//	gate.Record(syscall.SyscallMemoryAlloc, args, result, timestamp)
//	
//	// Get trace hash for MEG
//	traceHash := gate.Hash()
package syscall
