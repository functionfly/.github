package cert

import (
	"os"
	"strings"
	"testing"
)

func TestLoadAnchoringConfigFromEnv_Empty(t *testing.T) {
	for _, v := range []string{
		"ANCHOR_SIGNING_KEY", "ANCHOR_MIN_CONFIRMATIONS",
		"ANCHOR_RPC_BASE", "ANCHOR_RPC_ETHEREUM", "ANCHOR_RPC_POLYGON",
		"ANCHOR_RPC_ARBITRUM", "ANCHOR_RPC_OPTIMISM", "ANCHOR_RPC_AVALANCHE",
		"ANCHOR_CONTRACT_BASE", "ANCHOR_CONTRACT_ETHEREUM", "ANCHOR_CONTRACT_POLYGON",
		"ANCHOR_CONTRACT_ARBITRUM", "ANCHOR_CONTRACT_OPTIMISM", "ANCHOR_CONTRACT_AVALANCHE",
	} {
		os.Unsetenv(v)
	}

	cfg, err := LoadAnchoringConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IsEnabled() {
		t.Fatal("expected IsEnabled=false with no env vars set")
	}
	if cfg.SigningKeyHex != "" {
		t.Fatalf("expected empty signing key, got %q", cfg.SigningKeyHex)
	}
	if len(cfg.ConfiguredChains()) != 0 {
		t.Fatalf("expected no configured chains, got %v", cfg.ConfiguredChains())
	}
}

func TestLoadAnchoringConfigFromEnv_PartialRejected(t *testing.T) {
	os.Setenv("ANCHOR_RPC_BASE", "https://mainnet.base.org")
	defer os.Unsetenv("ANCHOR_RPC_BASE")

	cfg, err := LoadAnchoringConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for partial config (RPC without signing key)")
	}
}

func TestLoadAnchoringConfigFromEnv_Full(t *testing.T) {
	os.Setenv("ANCHOR_SIGNING_KEY", "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1")
	os.Setenv("ANCHOR_RPC_BASE", "https://mainnet.base.org")
	os.Setenv("ANCHOR_CONTRACT_BASE", "0x1234567890123456789012345678901234567890")
	os.Setenv("ANCHOR_MIN_CONFIRMATIONS", "12")
	defer func() {
		os.Unsetenv("ANCHOR_SIGNING_KEY")
		os.Unsetenv("ANCHOR_RPC_BASE")
		os.Unsetenv("ANCHOR_CONTRACT_BASE")
		os.Unsetenv("ANCHOR_MIN_CONFIRMATIONS")
	}()

	cfg, err := LoadAnchoringConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if !cfg.IsEnabled() {
		t.Fatal("expected IsEnabled=true")
	}
	chains := cfg.ConfiguredChains()
	if len(chains) != 1 || chains[0] != "base" {
		t.Fatalf("expected [base], got %v", chains)
	}
	if cfg.MinConfirmations != 12 {
		t.Fatalf("expected MinConfirmations=12, got %d", cfg.MinConfirmations)
	}
}

func TestAnchoringConfig_ValidateInvalidKey(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "not-a-hex-key",
		RPCEndpoints:       map[string]string{"base": "https://mainnet.base.org"},
		ContractAddresses:  map[string]string{"base": "0x1234567890123456789012345678901234567890"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for invalid signing key")
	}
}

func TestAnchoringConfig_ValidateInvalidContract(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:       map[string]string{"base": "https://mainnet.base.org"},
		ContractAddresses:  map[string]string{"base": "not-a-hex-address"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for invalid contract address")
	}
}

func TestAnchoringConfig_ValidateRPCWithoutContract(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:       map[string]string{"base": "https://mainnet.base.org"},
		ContractAddresses:  map[string]string{},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when RPC is set without contract")
	} else if !strings.Contains(err.Error(), "ANCHOR_CONTRACT_BASE") {
		t.Fatalf("expected error to mention ANCHOR_CONTRACT_BASE, got: %v", err)
	}
}

func TestAnchoringConfig_ValidateContractWithoutRPC(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:       map[string]string{},
		ContractAddresses:  map[string]string{"base": "0x1234567890123456789012345678901234567890"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when contract is set without RPC")
	} else if !strings.Contains(err.Error(), "ANCHOR_RPC_BASE") {
		t.Fatalf("expected error to mention ANCHOR_RPC_BASE, got: %v", err)
	}
}

func TestAnchoringConfig_ValidateUnsupportedChain(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:       map[string]string{"solana": "https://api.mainnet-beta.solana.com"},
		ContractAddresses:  map[string]string{"solana": "0x1234567890123456789012345678901234567890"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unsupported chain")
	}
}

func TestAnchoringConfig_ValidateInsecureRPC(t *testing.T) {
	cfg := &AnchoringConfig{
		SigningKeyHex:      "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:       map[string]string{"base": "ftp://mainnet.base.org"},
		ContractAddresses:  map[string]string{"base": "0x1234567890123456789012345678901234567890"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for non-http(s) RPC")
	}
}

func TestEthereumAnchoringService_Configure(t *testing.T) {
	svc := NewEthereumAnchoringService(nil)
	if svc.IsConfigured() {
		t.Fatal("expected IsConfigured=false on fresh service")
	}

	cfg := &AnchoringConfig{
		SigningKeyHex:     "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints:      map[string]string{"base": "https://mainnet.base.org"},
		ContractAddresses: map[string]string{"base": "0x1234567890123456789012345678901234567890"},
		MinConfirmations:  3,
	}
	if err := svc.Configure(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !svc.IsConfigured() {
		t.Fatal("expected IsConfigured=true after Configure")
	}
	chains := svc.Chains()
	if len(chains) != 1 || chains[0] != "base" {
		t.Fatalf("expected chains=[base], got %v", chains)
	}

	// Disable
	empty := &AnchoringConfig{}
	if err := svc.Configure(empty); err != nil {
		t.Fatalf("unexpected error on empty config: %v", err)
	}
	if svc.IsConfigured() {
		t.Fatal("expected IsConfigured=false after empty Configure")
	}
	if len(svc.Chains()) != 0 {
		t.Fatalf("expected empty chains after empty Configure, got %v", svc.Chains())
	}
}

func TestEthereumAnchoringService_Chains_Ordered(t *testing.T) {
	svc := NewEthereumAnchoringService(nil)
	cfg := &AnchoringConfig{
		SigningKeyHex: "0x4af1bceebf7f3634ec3cff8a2c38e51178d5d4ce585c52d6043cfe7f3b25d4e1",
		RPCEndpoints: map[string]string{
			"ethereum": "https://eth.llamarpc.com",
			"polygon":  "https://polygon-rpc.com",
			"base":     "https://mainnet.base.org",
		},
		ContractAddresses: map[string]string{
			"ethereum": "0x1234567890123456789012345678901234567890",
			"polygon":  "0x1234567890123456789012345678901234567890",
			"base":     "0x1234567890123456789012345678901234567890",
		},
	}
	if err := svc.Configure(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chains := svc.Chains()
	// Should be ordered by cost-efficiency: base, polygon, ethereum
	want := []string{"base", "polygon", "ethereum"}
	if len(chains) != len(want) {
		t.Fatalf("expected %d chains, got %d (%v)", len(want), len(chains), chains)
	}
	for i := range want {
		if chains[i] != want[i] {
			t.Fatalf("chain %d: expected %s, got %s", i, want[i], chains[i])
		}
	}
}
