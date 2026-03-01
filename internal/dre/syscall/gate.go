// Package syscall implements syscall virtualization for the DCC Protocol.
// It provides a gate (FX_SYSCALL_GATE) that controls which system calls are
// allowed in deterministic mode, and records syscalls for trace hashing.
package syscall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// SyscallID identifies the type of system call.
type SyscallID int

const (
	// Allowed syscalls (deterministic)
	SyscallMemoryAlloc    SyscallID = iota // Memory allocation
	SyscallMemoryRead                       // Virtual memory read
	SyscallMemoryWrite                      // Virtual memory write
	SyscallVirtFSRead                       // Virtual FS read
	SyscallVirtFSWrite                      // Virtual FS write
	SyscallClockRead                        // Virtual clock read
	SyscallRNG                              // Deterministic RNG
	SyscallLog                              // Structured logging
	SyscallEnvGet                           // Environment variable get
	SyscallArgGet                           // Command line argument get
	
	// Blocked syscalls (non-deterministic)
	SyscallSocketCreate    // Raw socket creation
	SyscallSystemClock     // Direct system clock
	SyscallDiskAccess      // Direct disk access
	SyscallProcessSpawn    // Process spawning
	SyscallHostFS          // Host file system access
	SyscallNetworkRaw      // Raw network operations
	SyscallRandom          // OS random source
	SyscallTime            // Wall time
	SyscallGetPID          // Process ID
	SyscallGetUID          // User ID
)

// SyscallGate implements the FX_SYSCALL_GATE for deterministic execution.
// It allows/blocks syscalls based on the syscall profile and records
// syscall traces for deterministic verification.
type SyscallGate struct {
	profile     string
	allowed     map[SyscallID]bool
	blocked     map[SyscallID]bool
	trace       []SyscallRecord
	traceHash   string
	mu          sync.RWMutex
	recordTrace bool
}

// SyscallRecord represents a single syscall for trace hashing.
type SyscallRecord struct {
	ID        SyscallID `json:"id"`
	Type      string    `json:"type"`
	Args      []byte    `json:"args"`
	Result    []byte    `json:"result"`
	Timestamp uint64    `json:"timestamp"`
}

// New creates a new SyscallGate with the given profile.
func New(profile string) *Gate {
	allowed := make(map[SyscallID]bool)
	blocked := make(map[SyscallID]bool)
	
	// Default allowed syscalls (deterministic)
	allowed[SyscallMemoryAlloc] = true
	allowed[SyscallMemoryRead] = true
	allowed[SyscallMemoryWrite] = true
	allowed[SyscallVirtFSRead] = true
	allowed[SyscallVirtFSWrite] = true
	allowed[SyscallClockRead] = true
	allowed[SyscallRNG] = true
	allowed[SyscallLog] = true
	allowed[SyscallEnvGet] = true
	allowed[SyscallArgGet] = true
	
	// Default blocked syscalls (non-deterministic)
	blocked[SyscallSocketCreate] = true
	blocked[SyscallSystemClock] = true
	blocked[SyscallDiskAccess] = true
	blocked[SyscallProcessSpawn] = true
	blocked[SyscallHostFS] = true
	blocked[SyscallNetworkRaw] = true
	blocked[SyscallRandom] = true
	blocked[SyscallTime] = true
	blocked[SyscallGetPID] = true
	blocked[SyscallGetUID] = true
	
	return &Gate{
		profile:     profile,
		allowed:    allowed,
		blocked:    blocked,
		trace:      make([]SyscallRecord, 0),
		recordTrace: true,
	}
}

// Allowed returns true if the syscall is allowed.
func (g *Gate) Allowed(id SyscallID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Check allowed list first
	if g.allowed[id] {
		return true
	}
	
	// Check blocked list
	if g.blocked[id] {
		return false
	}
	
	// Default to blocked for unknown syscalls
	return false
}

// Record records a syscall for trace hashing.
func (g *Gate) Record(id SyscallID, args []byte, result []byte, timestamp uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if !g.recordTrace {
		return
	}
	
	record := SyscallRecord{
		ID:        id,
		Type:      SyscallName(id),
		Args:      args,
		Result:    result,
		Timestamp: timestamp,
	}
	
	g.trace = append(g.trace, record)
}

// Hash returns the syscall trace hash.
// The hash is computed as: H("FX_SYSCALL" || syscall_type || canonical_args || result)
func (g *Gate) Hash() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if g.traceHash != "" {
		return g.traceHash
	}
	
	h := sha256.New()
	h.Write([]byte("FX_SYSCALL"))
	
	for _, record := range g.trace {
		h.Write([]byte(record.Type))
		if record.Args != nil {
			h.Write(record.Args)
		}
		if record.Result != nil {
			h.Write(record.Result)
		}
	}
	
	g.traceHash = hex.EncodeToString(h.Sum(nil))
	return g.traceHash
}

// EnableTrace enables syscall trace recording.
func (g *Gate) EnableTrace() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.recordTrace = true
}

// DisableTrace disables syscall trace recording.
func (g *Gate) DisableTrace() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.recordTrace = false
}

// Trace returns a copy of the syscall trace.
func (g *Gate) Trace() []SyscallRecord {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	trace := make([]SyscallRecord, len(g.trace))
	copy(trace, g.trace)
	
	return trace
}

// ClearTrace clears the syscall trace.
func (g *Gate) ClearTrace() {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.trace = make([]SyscallRecord, 0)
	g.traceHash = ""
}

// SyscallName returns the name of the syscall for logging.
func SyscallName(id SyscallID) string {
	names := map[SyscallID]string{
		SyscallMemoryAlloc:   "memory_alloc",
		SyscallMemoryRead:    "memory_read",
		SyscallMemoryWrite:   "memory_write",
		SyscallVirtFSRead:    "virt_fs_read",
		SyscallVirtFSWrite:   "virt_fs_write",
		SyscallClockRead:     "clock_read",
		SyscallRNG:           "rng",
		SyscallLog:           "log",
		SyscallEnvGet:        "env_get",
		SyscallArgGet:        "arg_get",
		SyscallSocketCreate:  "socket_create",
		SyscallSystemClock:   "system_clock",
		SyscallDiskAccess:    "disk_access",
		SyscallProcessSpawn:  "process_spawn",
		SyscallHostFS:        "host_fs",
		SyscallNetworkRaw:    "network_raw",
		SyscallRandom:        "random",
		SyscallTime:          "time",
		SyscallGetPID:        "get_pid",
		SyscallGetUID:        "get_uid",
	}
	
	if name, ok := names[id]; ok {
		return name
	}
	
	return fmt.Sprintf("unknown_syscall_%d", id)
}

// JSON returns the trace as JSON.
func (g *Gate) JSON() (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	data, err := json.MarshalIndent(g.trace, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal trace: %w", err)
	}
	
	return string(data), nil
}

// Gate is an alias for SyscallGate for convenience.
type Gate = SyscallGate
