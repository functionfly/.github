// Package clock implements the virtual clock for the DCC Protocol.
//
// The VirtualClock provides deterministic time progression independent of
// wall-clock time, ensuring reproducible execution across replays.
//
// Key properties:
//   - virtual_time = base_time + (tick_count * tick_size)
//   - base_time is derived from execution_id (deterministic)
//   - Time advances only on: explicit sleep, syscall, or tick
//   - No wall clock access allowed
//
// Usage:
//
//	execID := capsuleDescriptor.ExecutionID
//	clock := clock.New(execID)
//	
//	// Get current virtual time (in nanoseconds)
//	currentTime := clock.Now()
//	
//	// Advance time (e.g., on sleep syscall)
//	clock.Sleep(durationNanos)
//	
//	// Or tick per instruction batch
//	clock.Tick()
package clock
