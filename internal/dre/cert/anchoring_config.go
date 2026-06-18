package cert

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// AnchoringConfig holds the full configuration required to enable blockchain
// anchoring of execution certificates. All fields must be valid for the
// EthereumAnchoringService to operate; partial configuration is rejected so
// the service never silently no-ops on misconfiguration.
type AnchoringConfig struct {
	// SigningKeyHex is the ECDSA private key (hex, with or without 0x prefix).
	// In production this should be loaded from a secret manager (AWS Secrets
	// Manager, HashiCorp Vault, GCP Secret Manager) and never committed.
	SigningKeyHex string

	// RPCEndpoints maps chain name -> HTTPS RPC URL.
	RPCEndpoints map[string]string

	// ContractAddresses maps chain name -> anchoring contract address.
	// The contract must expose: function anchor(bytes32 root) external
	ContractAddresses map[string]string

	// MinConfirmations is the minimum number of block confirmations required
	// before an anchor is considered final. Defaults to 1 if zero.
	MinConfirmations int64
}

// LoadAnchoringConfigFromEnv builds an AnchoringConfig from environment
// variables. It returns a clear error listing every missing or invalid
// variable so operators can fix misconfiguration without guessing.
//
// Environment variables:
//
//	ANCHOR_SIGNING_KEY              - ECDSA private key, hex (0x prefix optional)
//	ANCHOR_MIN_CONFIRMATIONS        - (optional) int, default 1
//	ANCHOR_RPC_<CHAIN>              - HTTPS RPC URL for <CHAIN> (uppercase)
//	ANCHOR_CONTRACT_<CHAIN>         - Anchoring contract address for <CHAIN>
//
// Example for Base mainnet:
//
//	ANCHOR_SIGNING_KEY=0xabc...
//	ANCHOR_RPC_BASE=https://mainnet.base.org
//	ANCHOR_CONTRACT_BASE=0x123...
//	ANCHOR_MIN_CONFIRMATIONS=12
func LoadAnchoringConfigFromEnv() (*AnchoringConfig, error) {
	cfg := &AnchoringConfig{
		RPCEndpoints:      make(map[string]string),
		ContractAddresses: make(map[string]string),
		MinConfirmations:  1,
	}

	cfg.SigningKeyHex = strings.TrimSpace(os.Getenv("ANCHOR_SIGNING_KEY"))
	if cfg.SigningKeyHex != "" {
		cfg.SigningKeyHex = strings.TrimPrefix(cfg.SigningKeyHex, "0x")
	}

	if minConf := strings.TrimSpace(os.Getenv("ANCHOR_MIN_CONFIRMATIONS")); minConf != "" {
		var n int64
		if _, err := fmt.Sscanf(minConf, "%d", &n); err != nil || n < 0 {
			return nil, fmt.Errorf("cert: ANCHOR_MIN_CONFIRMATIONS must be a non-negative integer, got %q", minConf)
		}
		cfg.MinConfirmations = n
	}

	for _, chain := range SupportedChains {
		envChain := strings.ToUpper(chain)
		if rpc := strings.TrimSpace(os.Getenv("ANCHOR_RPC_" + envChain)); rpc != "" {
			cfg.RPCEndpoints[chain] = rpc
		}
		if addr := strings.TrimSpace(os.Getenv("ANCHOR_CONTRACT_" + envChain)); addr != "" {
			cfg.ContractAddresses[chain] = addr
		}
	}

	return cfg, nil
}

// Validate checks that the configuration is internally consistent and
// sufficient to perform anchoring. It does not perform network I/O.
//
// Rules:
//   - SigningKeyHex, if non-empty, must parse as a valid ECDSA private key.
//   - If any RPC endpoint or contract address is set, the signing key must
//     also be set (partial configuration is rejected).
//   - For each chain with an RPC endpoint, a contract address must be set
//     (and vice versa) — the service cannot anchor to a chain without both.
func (c *AnchoringConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("cert: nil anchoring config")
	}

	if c.SigningKeyHex != "" {
		key := strings.TrimPrefix(strings.TrimSpace(c.SigningKeyHex), "0x")
		if _, err := crypto.HexToECDSA(key); err != nil {
			return fmt.Errorf("cert: invalid ANCHOR_SIGNING_KEY: %w", err)
		}
	}

	hasAnyRPC := len(c.RPCEndpoints) > 0
	hasAnyContract := len(c.ContractAddresses) > 0

	if c.SigningKeyHex == "" && (hasAnyRPC || hasAnyContract) {
		return fmt.Errorf("cert: ANCHOR_RPC_* / ANCHOR_CONTRACT_* set but ANCHOR_SIGNING_KEY is empty; refusing partial configuration")
	}

	if c.SigningKeyHex == "" {
		return nil
	}

	chainsWithRPC := make(map[string]struct{}, len(c.RPCEndpoints))
	for chain, rpc := range c.RPCEndpoints {
		if !IsChainSupported(chain) {
			return fmt.Errorf("cert: ANCHOR_RPC_%s references unsupported chain %q", strings.ToUpper(chain), chain)
		}
		if !strings.HasPrefix(rpc, "https://") && !strings.HasPrefix(rpc, "http://") {
			return fmt.Errorf("cert: ANCHOR_RPC_%s must be an http(s) URL, got %q", strings.ToUpper(chain), rpc)
		}
		chainsWithRPC[chain] = struct{}{}
		if _, ok := c.ContractAddresses[chain]; !ok {
			return fmt.Errorf("cert: ANCHOR_RPC_%s is set but ANCHOR_CONTRACT_%s is missing", strings.ToUpper(chain), strings.ToUpper(chain))
		}
	}

	chainsWithContract := make(map[string]struct{}, len(c.ContractAddresses))
	for chain, addr := range c.ContractAddresses {
		if !IsChainSupported(chain) {
			return fmt.Errorf("cert: ANCHOR_CONTRACT_%s references unsupported chain %q", strings.ToUpper(chain), chain)
		}
		if !common.IsHexAddress(addr) {
			return fmt.Errorf("cert: ANCHOR_CONTRACT_%s is not a valid hex address: %q", strings.ToUpper(chain), addr)
		}
		chainsWithContract[chain] = struct{}{}
		if _, ok := c.RPCEndpoints[chain]; !ok {
			return fmt.Errorf("cert: ANCHOR_CONTRACT_%s is set but ANCHOR_RPC_%s is missing", strings.ToUpper(chain), strings.ToUpper(chain))
		}
	}

	if len(chainsWithRPC) == 0 {
		return fmt.Errorf("cert: ANCHOR_SIGNING_KEY is set but no ANCHOR_RPC_<CHAIN> / ANCHOR_CONTRACT_<CHAIN> pairs are configured")
	}

	return nil
}

// PublicKey returns the ECDSA public key derived from the configured signing key.
// Returns nil if the key is not set or invalid.
func (c *AnchoringConfig) PublicKey() (*ecdsa.PublicKey, error) {
	if c == nil || c.SigningKeyHex == "" {
		return nil, fmt.Errorf("cert: no signing key configured")
	}
	key, err := crypto.HexToECDSA(c.SigningKeyHex)
	if err != nil {
		return nil, fmt.Errorf("cert: invalid signing key: %w", err)
	}
	return &key.PublicKey, nil
}

// ConfiguredChains returns the list of chains that have both an RPC endpoint
// and a contract address configured.
func (c *AnchoringConfig) ConfiguredChains() []string {
	if c == nil {
		return nil
	}
	out := make([]string, 0, len(c.RPCEndpoints))
	for _, chain := range SupportedChains {
		if _, okRPC := c.RPCEndpoints[chain]; okRPC {
			if _, okContract := c.ContractAddresses[chain]; okContract {
				out = append(out, chain)
			}
		}
	}
	return out
}

// IsEnabled reports whether anchoring is fully configured and ready to use.
func (c *AnchoringConfig) IsEnabled() bool {
	if c == nil {
		return false
	}
	return c.SigningKeyHex != "" && len(c.ConfiguredChains()) > 0
}

// Configure applies a validated AnchoringConfig to the EthereumAnchoringService.
// It is safe to call once at startup; subsequent calls replace the previous
// configuration. The signing key is zeroed in the returned error path is not
// retained beyond validation.
func (s *EthereumAnchoringService) Configure(cfg *AnchoringConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.SigningKeyHex == "" {
		// Explicitly disable anchoring.
		s.privateKey = ""
		s.rpcEndpoints = map[string]string{}
		s.contractAddresses = map[string]string{}
		return nil
	}

	rpcs := make(map[string]string, len(cfg.RPCEndpoints))
	for k, v := range cfg.RPCEndpoints {
		rpcs[k] = v
	}
	contracts := make(map[string]string, len(cfg.ContractAddresses))
	for k, v := range cfg.ContractAddresses {
		contracts[k] = v
	}

	s.privateKey = strings.TrimPrefix(strings.TrimSpace(cfg.SigningKeyHex), "0x")
	s.rpcEndpoints = rpcs
	s.contractAddresses = contracts

	// Drop cached clients so the next Anchor() dials with the new endpoints.
	s.clientMu.Lock()
	s.clientCache = nil
	s.clientMu.Unlock()

	return nil
}
