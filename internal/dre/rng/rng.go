// Package rng implements deterministic random number generation for the DCC Protocol.
// It uses ChaCha20 with a seed derived from the capsule's RNGSeed to ensure
// deterministic, reproducible results across executions and replays.
package rng

import (
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/hkdf"
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
func New(seed []byte) (*DeterministicRNG, error) {
	// ChaCha20 requires a 32-byte key and 12-byte nonce
	key, err := deriveKey(seed)
	if err != nil {
		return nil, err
	}
	nonce, err := deriveNonce(seed)
	if err != nil {
		return nil, err
	}

	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		// This should never happen with valid inputs
		return nil, fmt.Errorf("rng: failed to create ChaCha20 cipher: %w", err)
	}

	return &DeterministicRNG{
		cipher:  c,
		seed:    seed,
		counter: 0,
	}, nil
}

// Seed reinitializes the RNG with a new seed (for fresh execution, not replay).
func (r *DeterministicRNG) Seed(seed []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, err := deriveKey(seed)
	if err != nil {
		return err
	}
	nonce, err := deriveNonce(seed)
	if err != nil {
		return err
	}

	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return fmt.Errorf("rng: failed to reinitialize ChaCha20 cipher: %w", err)
	}

	r.cipher = c
	r.seed = seed
	r.counter = 0
	return nil
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

// DCC domain separation for HKDF (fixed per protocol).
var kdfSalt = []byte("DCC-RNG-v1")

// deriveKey derives a 32-byte ChaCha20 key from the seed using HKDF-SHA256.
// Deterministic and suitable for production; same seed always yields the same key.
func deriveKey(seed []byte) ([]byte, error) {
	h := hkdf.New(sha256.New, seed, kdfSalt, []byte("rng-key"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, fmt.Errorf("rng: deriveKey: %w", err)
	}
	return key, nil
}

// deriveNonce derives a 12-byte ChaCha20 nonce from the seed using HKDF-SHA256.
// Uses different info than deriveKey to avoid correlation.
func deriveNonce(seed []byte) ([]byte, error) {
	h := hkdf.New(sha256.New, seed, kdfSalt, []byte("rng-nonce"))
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(h, nonce); err != nil {
		return nil, fmt.Errorf("rng: deriveNonce: %w", err)
	}
	return nonce, nil
}
