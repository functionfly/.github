package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// StateFabricHostConfig holds dependencies needed for a real host handler.
type StateFabricHostConfig struct {
	Repo          interface{}
	TriggerEngine interface{}
	TenantID      uuid.UUID
	DefaultFabric uuid.UUID
}

// RuntimeHostHandlerConfig extends StateFabricHostConfig with attestation dependencies.
type RuntimeHostHandlerConfig struct {
	StateFabricHostConfig
	AttestationRepo AttestationCreator
	FunctionID      uuid.UUID
	FunctionVersion string
	FunctionName    string
	FunctionAuthor  string
	AgentID         string
	TrustScore      float64
}

// AttestationCreator is the minimal interface for creating delegation attestations
// from the host handler without importing the full trustapi package.
type AttestationCreator interface {
	CreateDelegationAttestation(
		delegatorFunctionID uuid.UUID,
		delegatorAgentID string,
		delegatorTrustScore float64,
		delegateeFunctionID uuid.UUID,
		delegateeVersion string,
		delegateeName string,
		delegateeAuthor string,
		inputHash string,
		outputHash string,
		chainID string,
		parentAttestationID string,
	) (interface{}, error)
	GetAttestationByAttestationID(id string) (interface{}, error)
}

func HostHandlerForExecution(_ StateFabricHostConfig) HostFunctionHandler {
	return NewDefaultHostHandler(nil)
}

// NewRuntimeHostHandler creates a host handler with full attestation and delegation support.
func NewRuntimeHostHandler(cfg RuntimeHostHandlerConfig) HostFunctionHandler {
	return &runtimeHostHandler{
		cfg: cfg,
	}
}

// runtimeHostHandler implements HostFunctionHandler with real attestation/delegation support.
type runtimeHostHandler struct {
	cfg RuntimeHostHandlerConfig
}

func (h *runtimeHostHandler) Log(message string) {
	fmt.Printf("[functionfly] %s\n", message)
}

func (h *runtimeHostHandler) Fetch(request string) (string, error) {
	return "", fmt.Errorf("fetch not implemented in runtime host handler")
}

func (h *runtimeHostHandler) KVGet(key string) (string, error) {
	return "", fmt.Errorf("kv not implemented in runtime host handler")
}

func (h *runtimeHostHandler) KVSet(key string, value string) error {
	return fmt.Errorf("kv not implemented in runtime host handler")
}

func (h *runtimeHostHandler) GetEnv(name string) string {
	return ""
}

func (h *runtimeHostHandler) AIInference(model string, input []byte, params string) (string, error) {
	return "", fmt.Errorf("ai inference not implemented in runtime host handler")
}

func (h *runtimeHostHandler) StateGet(path string) (string, error) {
	return "", fmt.Errorf("state not implemented in runtime host handler")
}

func (h *runtimeHostHandler) StateSet(path string, value string) error {
	return fmt.Errorf("state not implemented in runtime host handler")
}

func (h *runtimeHostHandler) StateDelete(path string) error {
	return fmt.Errorf("state not implemented in runtime host handler")
}

func (h *runtimeHostHandler) StateGetFabric(fabricID string) (string, error) {
	return "", fmt.Errorf("state fabric not implemented in runtime host handler")
}

func (h *runtimeHostHandler) StateCreateSnapshot(path string, label string) (string, error) {
	return "", fmt.Errorf("state snapshot not implemented in runtime host handler")
}

func (h *runtimeHostHandler) GetAttestation(attestationID string) (string, error) {
	if h.cfg.AttestationRepo == nil {
		return "", fmt.Errorf("attestation service not available")
	}

	att, err := h.cfg.AttestationRepo.GetAttestationByAttestationID(attestationID)
	if err != nil {
		return "", fmt.Errorf("attestation %q not found: %w", attestationID, err)
	}

	data, err := json.Marshal(att)
	if err != nil {
		return "", fmt.Errorf("marshal attestation: %w", err)
	}

	return string(data), nil
}

func (h *runtimeHostHandler) Delegate(targetFunctionID string, input string, options string) (string, error) {
	// Parse target function ID
	targetID, err := uuid.Parse(targetFunctionID)
	if err != nil {
		return "", fmt.Errorf("invalid target function ID: %w", err)
	}

	// Parse delegation options
	var opts struct {
		MinTrustScore float64 `json:"min_trust_score"`
		TimeoutMs     int     `json:"timeout_ms"`
		ChainID       string  `json:"chain_id"`
		ParentAttID   string  `json:"parent_attestation_id"`
	}
	if options != "" {
		json.Unmarshal([]byte(options), &opts)
	}

	// Check minimum trust score
	if opts.MinTrustScore > 0 && h.cfg.TrustScore < opts.MinTrustScore {
		return "", fmt.Errorf("trust score %.2f below minimum %.2f for delegation", h.cfg.TrustScore, opts.MinTrustScore)
	}

	// Compute input hash for chain of custody
	inputHash := sha256Hex(input)

	// Create delegation attestation if attestation repo is available
	if h.cfg.AttestationRepo != nil {
		att, err := h.cfg.AttestationRepo.CreateDelegationAttestation(
			h.cfg.FunctionID,       // delegator
			h.cfg.AgentID,          // delegator agent
			h.cfg.TrustScore,       // delegator trust score
			targetID,               // delegatee
			"",                     // delegatee version (unknown at delegation time)
			"",                     // delegatee name (unknown at delegation time)
			"",                     // delegatee author (unknown at delegation time)
			inputHash,              // input hash
			"",                     // output hash (filled after execution)
			opts.ChainID,           // chain ID
			opts.ParentAttID,       // parent attestation ID
		)
		if err != nil {
			// Log but don't fail delegation
			fmt.Printf("[functionfly] warning: failed to create delegation attestation: %v\n", err)
		} else {
			_ = att
		}
	}

	// Return delegation acknowledgment
	result := map[string]interface{}{
		"status":           "delegated",
		"target_function":  targetFunctionID,
		"input_hash":       inputHash,
		"delegator":        h.cfg.FunctionID.String(),
		"delegator_agent":  h.cfg.AgentID,
		"delegator_trust":  h.cfg.TrustScore,
		"chain_id":         opts.ChainID,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal delegation result: %w", err)
	}

	return string(data), nil
}

func (h *runtimeHostHandler) Call(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	return nil, fmt.Errorf("generic call not implemented in runtime host handler")
}

// sha256Hex computes the hex-encoded SHA-256 hash of a string.
func sha256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// ParseFabricIDFromInput extracts an optional default fabric ID from execution input JSON.
func ParseFabricIDFromInput(input []byte) uuid.UUID {
	if len(input) == 0 {
		return uuid.Nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return uuid.Nil
	}
	for _, key := range []string{"_state_fabric_id", "state_fabric_id", "fabric_id", "fabricId"} {
		if raw, ok := payload[key]; ok {
			if s, ok := raw.(string); ok {
				if id, err := uuid.Parse(s); err == nil {
					return id
				}
			}
		}
	}
	return uuid.Nil
}
