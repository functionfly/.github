// Package rng implements deterministic random number generation for the DCC Protocol.
// It uses ChaCha20 with a seed derived from the capsule's RNGSeed to ensure
// deterministic, reproducible results across executions and replays.
package rng

import (
	"crypto/cipher"
	"crypto/chacha20"
	"encoding/binary"
	"sync"
)

// DeterministicRNG is a ChaCha20-based deterministic random number generator.
// The seed is derived from the capsule's RNGSeed and remains constant across replays.
type DeterministicRNG struct {
	cipher  cipher.Stream
	seed    []byte
	counter uint64
	mu      sync.Mutex
}

// New creates a new DeterministicRNG with the given seed.
// The seed should be derived from H(InputHash || EnvironmentHash) per DCC spec.
func New(seed []byte) *DeterministicRNG {
	// ChaCha20 requires a 32-byte key and 12-byte nonce
	key := deriveKey(seed)
	nonce := deriveNonce(seed)
	
	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		// This should never happen with valid inputs
		panic("rng: failed to create ChaCha20 cipher: " + err.Error())
	}
	
	return &DeterministicRNG{
		cipher: c,
		seed:   seed,
		counter: 0,
	}
}

// Seed reinitializes the RNG with a new seed (for fresh execution, not replay).
func (r *DeterministicRNG) Seed(seed []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := deriveKey(seed)
	nonce := deriveNonce(seed)
	
	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		panic("rng: failed to reinitialize ChaCha20 cipher: " + err.Error())
	}
	
	r.cipher = c
	r.seed = seed
	r.counter = 0
}

// Next returns the next deterministic uint32 value.
func (r *DeterministicRNG) Next() uint32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var buf [4]byte
	r.cipher.XORKeyStream(buf[:], buf[:])
	r.counter++
	return binary.LittleEndian.Uint32(buf[:])
}

// Next64 returns the next deterministic uint64 value.
func (r *DeterministicRNG) Next64() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var buf [8]byte
	r.cipher.XORKeyStream(buf[:], buf[:])
	r.counter++
	return binary.LittleEndian.Uint64(buf[:])
}

// Read fills the given slice with deterministic random bytes.
// It returns the number of bytes written.
func (r *DeterministicRNG) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if len(p) == 0 {
		return 0, nil
	}
	
	r.cipher.XORKeyStream(p, p)
	r.counter++
	return len(p), nil
}

// Counter returns the current call count (number of random operations performed).
func (r *DeterministicRNG) Counter() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counter
}

// Intn returns a deterministic integer in [0, n).
func (r *DeterministicRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.Next64() % uint64(n))
}

// Int31n returns a deterministic int31 in [0, n).
func (r *DeterministicRNG) Int31n(n int32) int32 {
	if n <= 0 {
		return 0
	}
	return int32(r.Next() % uint32(n))
}

// Float64 returns a deterministic float64 in [0, 1).
func (r *DeterministicRNG) Float64() float64 {
	// Use the algorithm from math/rand for consistent results
	return float64(r.Next64()&(1<<53-1)) / float64(1<<53)
}

// deriveKey derives a 32-byte key from the seed using simple hashing.
// In production, this should use a proper KDF, but for DCC determinism,
// we just need a deterministic derivation.
func deriveKey(seed []byte) []byte {
	// Simple derivation: repeat or hash seed to 32 bytes
	key := make([]byte, 32)
	seedLen := len(seed)
	
	for i := 0; i < 32; i++ {
		key[i] = seed[i%seedLen]
	}
	
	return key
}

// deriveNonce derives a 12-byte nonce from the seed.
func deriveNonce(seed []byte) []byte {
	nonce := make([]byte, 12)
	seedLen := len(seed)
	
	// Use different derivation than key to avoid correlation
	offset := seedLen / 2
	for i := 0; i < 12; i++ {
		nonce[i] = seed[(i+offset)%seedLen]
	}
	
	return nonce
}
