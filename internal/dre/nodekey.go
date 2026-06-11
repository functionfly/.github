// Package dre provides DRE (Deterministic Reliable Execution) protocol support.
// This file loads the execution node's Ed25519 key from environment for signing FXCERTs.
//
// SECURITY WARNING: The private key is loaded from environment variables or files.
// - Never log the private key or include it in error messages
// - Ensure the key file has restricted permissions (600 or similar)
// - PRODUCTION: Use HSM (AWS KMS, GCP Cloud KMS, HashiCorp Vault) for key management
// - When keys are loaded, they are held in memory - ensure secure memory handling in production
// - For production deployments, prefer DRE_PLATFORM_PRIVATE_KEY_PATH pointing to a secrets manager mount

package dre

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"strings"
)

// Env names for DRE node configuration (optional; when set, FXCERTs are signed).
const (
	EnvNodePrivateKey     = "DRE_NODE_PRIVATE_KEY"      // base64-encoded 64-byte Ed25519 private key
	EnvNodePrivateKeyPath = "DRE_NODE_PRIVATE_KEY_PATH" // path to file: raw 64 bytes or base64
	EnvNodeID             = "DRE_NODE_ID"               // node identifier (default "bootstrap")
	EnvRegion             = "DRE_REGION"                // region (default "internal")

	EnvPlatformPrivateKey     = "DRE_PLATFORM_PRIVATE_KEY"      // base64 64-byte Ed25519 platform key (optional)
	EnvPlatformPrivateKeyPath = "DRE_PLATFORM_PRIVATE_KEY_PATH" // or path to file
)

// LoadNodeKeyFromEnv loads the DRE node Ed25519 private key and optional node ID/region from environment.
// If no key is configured, returns (nil, nodeID, region, nil) so callers can still use node ID/region.
// Key can be provided as:
//   - DRE_NODE_PRIVATE_KEY: base64-encoded 64-byte private key (Go ed25519 format)
//   - DRE_NODE_PRIVATE_KEY_PATH: path to file containing raw 64 bytes or base64 (one line)
func LoadNodeKeyFromEnv() (key ed25519.PrivateKey, nodeID, region string, err error) {
	nodeID = strings.TrimSpace(os.Getenv(EnvNodeID))
	if nodeID == "" {
		nodeID = "bootstrap"
	}
	region = strings.TrimSpace(os.Getenv(EnvRegion))
	if region == "" {
		region = "internal"
	}

	// Try inline base64 first
	if b64 := strings.TrimSpace(os.Getenv(EnvNodePrivateKey)); b64 != "" {
		key, err = decodeEd25519PrivateKey(b64)
		return key, nodeID, region, err
	}

	// Try file path
	if path := strings.TrimSpace(os.Getenv(EnvNodePrivateKeyPath)); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nodeID, region, err
		}
		key, err = decodeEd25519PrivateKeyBytes(data)
		return key, nodeID, region, err
	}

	return nil, nodeID, region, nil
}

// LoadPlatformKeyFromEnv loads the optional DRE platform Ed25519 private key from environment.
// When set, FXCERTs include a platform signature and Platform Key ID is shown in the UI.
// Uses DRE_PLATFORM_PRIVATE_KEY (base64) or DRE_PLATFORM_PRIVATE_KEY_PATH (file).
func LoadPlatformKeyFromEnv() (ed25519.PrivateKey, error) {
	if b64 := strings.TrimSpace(os.Getenv(EnvPlatformPrivateKey)); b64 != "" {
		return decodeEd25519PrivateKey(b64)
	}
	if path := strings.TrimSpace(os.Getenv(EnvPlatformPrivateKeyPath)); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return decodeEd25519PrivateKeyBytes(data)
	}
	return nil, nil
}

// decodeEd25519PrivateKey decodes a base64 string (e.g. from DRE_NODE_PRIVATE_KEY) into an Ed25519 private key.
func decodeEd25519PrivateKey(s string) (ed25519.PrivateKey, error) {
	s = strings.TrimSpace(s)
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return nil, nil
	}
	return ed25519.PrivateKey(decoded), nil
}

// decodeEd25519PrivateKeyBytes decodes file content: raw 64 bytes or base64.
func decodeEd25519PrivateKeyBytes(data []byte) (ed25519.PrivateKey, error) {
	data = trimBytes(data)
	if len(data) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(data), nil
	}
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil || len(decoded) != ed25519.PrivateKeySize {
		return nil, nil
	}
	return ed25519.PrivateKey(decoded), nil
}

func trimBytes(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}
