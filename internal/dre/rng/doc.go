// Package rng implements deterministic random number generation for the DCC Protocol.
//
// The DeterministicRNG uses ChaCha20 with a seed derived from the capsule's
// RNGSeed (H(InputHash || EnvironmentHash)) to ensure deterministic,
// reproducible results across executions and replays.
//
// Key properties:
//   - No hardware or OS entropy is used
//   - Counter strictly increments per call
//   - Thread-safe for potential future multi-threading
//   - Same seed always produces identical sequences
//
// Usage:
//
//	seed := capsuleDescriptor.RNGSeed // H(input_hash || env_hash)
//	rng := rng.New([]byte(seed))
//	
//	// Deterministic random values
//	val := rng.Next()       // uint32
//	val64 := rng.Next64()   // uint64
//	rand := rng.Float64()   // float64 in [0, 1)
//	index := rng.Intn(100)  // int in [0, 100)
package rng
