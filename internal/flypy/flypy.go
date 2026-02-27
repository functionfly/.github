package flypy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/flypy/artifact"
	"github.com/functionfly/functionfly/internal/flypy/backend"
	"github.com/functionfly/functionfly/internal/flypy/compiler"
	"github.com/functionfly/functionfly/internal/flypy/ir"
	"github.com/functionfly/functionfly/internal/flypy/parser"
	"github.com/functionfly/functionfly/internal/flypy/restrictions"
	"github.com/functionfly/functionfly/internal/flypy/verifier"
)

// Config holds FlyPy configuration options
type Config struct {
	// Mode determines the execution mode
	Mode ExecutionMode

	// OutputDir is where compiled artifacts are written
	OutputDir string

	// Verbose enables detailed logging
	Verbose bool

	// SignKey is the Ed25519 private key for signing artifacts
	SignKey []byte

	// TargetWasm specifies the Wasm target (default: wasm32-unknown-unknown)
	TargetWasm string

	// Version specifies the function version (default: 1.0.0)
	Version string
}

// ExecutionMode defines how the function will be executed
type ExecutionMode string

const (
	// DeterministicMode compiles to pure Wasm with full determinism
	// Only allows: json, math, typing, collections
	DeterministicMode ExecutionMode = "deterministic"

	// ComplexMode allows extended stdlib modules while maintaining determinism
	// Allows: csv, io (StringIO/BytesIO), re, datetime, itertools, functools, etc.
	ComplexMode ExecutionMode = "complex"

	// CompatibleMode allows some non-deterministic operations (with warnings)
	// Uses MicroPython fallback for full Python compatibility
	CompatibleMode ExecutionMode = "compatible"
)

// Result contains the result of a FlyPy compilation
type Result struct {
	// Artifact is the compiled and signed artifact bundle
	Artifact *artifact.Artifact

	// Warnings contains any warnings encountered during compilation
	Warnings []string

	// DeterminismProof contains proof of determinism for verification
	DeterminismProof *DeterminismProof

	// SideEffects contains the side effect analysis results
	SideEffects []verifier.SideEffect

	// SideEffectSummary provides a summary of side effects by type
	SideEffectSummary map[verifier.SideEffectType]int
}

// DeterminismProof contains cryptographic proof of determinism
type DeterminismProof struct {
	// IRHash is the SHA-256 hash of the canonical IR
	IRHash string

	// Timestamp is when the proof was generated
	Timestamp string

	// Capabilities are the declared capabilities
	Capabilities []string
}

// Compiler is the main FlyPy compiler
type Compiler struct {
	config *Config
}

// NewCompiler creates a new FlyPy compiler with the given configuration
func NewCompiler(config *Config) *Compiler {
	if config.OutputDir == "" {
		config.OutputDir = "./dist"
	}
	if config.TargetWasm == "" {
		config.TargetWasm = "wasm32-unknown-unknown"
	}
	return &Compiler{
		config: config,
	}
}

// NewCompilerWithDefaults creates a new FlyPy compiler with default configuration
func NewCompilerWithDefaults() *Compiler {
	return NewCompiler(&Config{
		Mode:      DeterministicMode,
		OutputDir: "./dist",
		Verbose:   false,
	})
}

// Compile compiles Python source code to a deterministic Wasm artifact
func (c *Compiler) Compile(ctx context.Context, source string, name string) (*Result, error) {
	if c.config.Verbose {
		fmt.Println("🔨 Compiling", name, "to deterministic Wasm...")
	}

	// Phase 1: Parse Python source to AST
	if c.config.Verbose {
		fmt.Println("✓ Parsing Python AST")
	}
	pythonAST, err := parser.ParsePython(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Python: %w", err)
	}

	// Phase 2: Enforce restricted subset (mode-aware)
	if c.config.Verbose {
		fmt.Println("✓ Enforcing restricted subset")
	}
	restrictionErrors := restrictions.EnforceWithMode(pythonAST, restrictions.ExecutionMode(c.config.Mode))
	if len(restrictionErrors) > 0 {
		return nil, fmt.Errorf("restriction violations: %v", restrictionErrors)
	}

	// Phase 3: Generate IR
	if c.config.Verbose {
		fmt.Println("✓ Generating deterministic IR")
	}
	irModule, err := ir.Generate(pythonAST, name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate IR: %w", err)
	}

	// Phase 4: Verify determinism
	if c.config.Verbose {
		fmt.Println("✓ Verifying determinism")
	}
	verificationErrors := verifier.Verify(irModule)
	if len(verificationErrors) > 0 {
		return nil, fmt.Errorf("determinism verification failed: %v", verificationErrors)
	}

	// Phase 4.5: Analyze side effects
	if c.config.Verbose {
		fmt.Println("✓ Analyzing side effects")
	}
	sideEffectAnalyzer := verifier.NewSideEffectAnalyzer(irModule)
	sideEffects := sideEffectAnalyzer.Analyze()

	// Check for side effects that violate determinism
	for _, effect := range sideEffects {
		if effect.Type == verifier.SideEffectNetwork ||
			effect.Type == verifier.SideEffectExternalState ||
			effect.Type == verifier.SideEffectIO {
			return nil, fmt.Errorf("side effect violation: %s in function %s", effect.Message, effect.Function)
		}
	}

	// Phase 5: Generate Rust code (mode-aware)
	if c.config.Verbose {
		fmt.Println("✓ Generating Rust code")
	}
	rustCode, err := backend.GenerateRustWithMode(irModule, string(c.config.Mode))
	if err != nil {
		return nil, fmt.Errorf("failed to generate Rust: %w", err)
	}

	// Phase 6: Compile to Wasm
	if c.config.Verbose {
		fmt.Println("✓ Compiling to Wasm")
	}
	wasmBytes, err := compiler.CompileRustWithMode(rustCode, c.config.TargetWasm, string(c.config.Mode))
	if err != nil {
		return nil, fmt.Errorf("failed to compile Wasm: %w", err)
	}

	// Phase 7: Build artifact bundle
	if c.config.Verbose {
		fmt.Println("✓ Building artifact bundle")
	}
	version := c.config.Version
	if version == "" {
		version = "1.0.0"
	}

	artifactBundle, err := artifact.Build(artifact.BuildInput{
		WasmModule:    wasmBytes,
		IRModule:      irModule,
		Name:          name,
		Version:       version,
		SignKey:       c.config.SignKey,
		Deterministic: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build artifact: %w", err)
	}

	// Write output files
	if err := c.writeOutput(artifactBundle); err != nil {
		return nil, fmt.Errorf("failed to write output: %w", err)
	}

	if c.config.Verbose {
		fmt.Println("✅ Build complete:", c.config.OutputDir)
	}

	return &Result{
		Artifact: artifactBundle,
		Warnings: []string{},
		DeterminismProof: &DeterminismProof{
			IRHash:       artifactBundle.DeterminismHash,
			Capabilities: artifactBundle.CapabilityMap.Requested,
		},
		SideEffects:       sideEffects,
		SideEffectSummary: sideEffectAnalyzer.GetSideEffectSummary(),
	}, nil
}

// writeOutput writes the artifact bundle to the output directory
func (c *Compiler) writeOutput(artifact *artifact.Artifact) error {
	// Create output directory
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return err
	}

	// Write Wasm module
	wasmPath := filepath.Join(c.config.OutputDir, "state_transition.wasm")
	if err := os.WriteFile(wasmPath, artifact.WasmModule, 0644); err != nil {
		return err
	}

	// Write manifest
	manifestPath := filepath.Join(c.config.OutputDir, "manifest.json")
	manifestJSON, err := json.MarshalIndent(artifact.Manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return err
	}

	// Write capability map
	capPath := filepath.Join(c.config.OutputDir, "capability.map")
	capJSON, err := json.MarshalIndent(artifact.CapabilityMap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(capPath, capJSON, 0644); err != nil {
		return err
	}

	// Write determinism hash
	hashPath := filepath.Join(c.config.OutputDir, "determinism.hash")
	if err := os.WriteFile(hashPath, []byte(artifact.DeterminismHash), 0644); err != nil {
		return err
	}

	// Write signature
	sigPath := filepath.Join(c.config.OutputDir, "signature.sig")
	if err := os.WriteFile(sigPath, artifact.Signature, 0644); err != nil {
		return err
	}

	return nil
}

// GetVersion returns the current FlyPy version
func GetVersion() string {
	return "1.0.0"
}

// SupportsLanguage checks if a language is supported for deterministic compilation
func SupportsLanguage(lang string) bool {
	switch strings.ToLower(lang) {
	case "python", "python3", "py":
		return true
	default:
		return false
	}
}
