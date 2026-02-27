// Package main implements the FlyPy Python-to-WASM compiler CLI.
// This compiler transforms Python functions into deterministic WebAssembly modules
// that execute in the FunctionFly runtime without requiring a Python interpreter.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/flypy"
	"github.com/functionfly/functionfly/internal/flypy/verifier"
)

// CompileFlags represents the command-line flags for the compiler
type CompileFlags struct {
	input             string
	metadata          string
	output            string
	mode              string
	optimize          string
	optimizeColdStart bool
	verbose           bool
	signingKey        string
	timeout           time.Duration
}

// Metadata represents function metadata from the input JSON file
type Metadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	EntryPoint  string `json:"entry_point"`
	Runtime     string `json:"runtime"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
}

// CompileError represents a structured compilation error
type CompileError struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Source  string `json:"source,omitempty"`
}

func (e *CompileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] %s at line %d, column %d", e.Phase, e.Message, e.Line, e.Column)
	}
	return fmt.Sprintf("[%s] %s", e.Phase, e.Message)
}

// CompilationResult represents the result of compilation
type CompilationResult struct {
	Success         bool                            `json:"success"`
	WASMBytes       []byte                          `json:"-"`
	Manifest        map[string]interface{}          `json:"manifest"`
	Errors          []CompileError                  `json:"errors,omitempty"`
	Warnings        []string                        `json:"warnings,omitempty"`
	SideEffects     []verifier.SideEffect           `json:"side_effects,omitempty"`
	SideEffectStats map[verifier.SideEffectType]int `json:"side_effect_stats,omitempty"`
}

// Valid compilation modes
const (
	ModeDeterministic = "deterministic"
	ModeComplex       = "complex"
	ModeCompatible    = "compatible"
)

// Valid optimization levels
const (
	OptMinimal    = "minimal"
	OptBalanced   = "balanced"
	OptAggressive = "aggressive"
)

func runCompile(args []string) error {
	flags, err := parseAndValidateFlags(args)
	if err != nil {
		printCompileHelp()
		return err
	}

	// Create output directory early
	if err := os.MkdirAll(flags.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read and validate input files
	sourceCode, metadata, err := readInputFiles(flags)
	if err != nil {
		return err
	}

	// Validate source code
	if err := validateSourceCode(sourceCode); err != nil {
		return err
	}

	if flags.verbose {
		printCompilationHeader(flags, metadata, len(sourceCode))
	}

	// Compile Python to WASM
	result, err := compilePythonToWASM(sourceCode, metadata, flags)
	if err != nil {
		return err
	}

	if !result.Success {
		printCompilationErrors(result)
		return errors.New("compilation failed")
	}

	// Write output files
	if err := writeOutputFiles(flags, result, *metadata); err != nil {
		return fmt.Errorf("failed to write output files: %w", err)
	}

	printCompilationSuccess(result, *metadata, flags.verbose)
	return nil
}

// parseAndValidateFlags parses command-line flags and validates them
func parseAndValidateFlags(args []string) (*CompileFlags, error) {
	flags := &CompileFlags{
		timeout: 5 * time.Minute,
	}

	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.StringVar(&flags.input, "input", "", "Input Python source file (required)")
	fs.StringVar(&flags.metadata, "metadata", "", "Input metadata JSON file (required)")
	fs.StringVar(&flags.output, "output", "", "Output directory (required)")
	fs.StringVar(&flags.mode, "mode", ModeDeterministic, "Execution mode: deterministic, complex, or compatible")
	fs.StringVar(&flags.optimize, "optimize", OptBalanced, "Optimization level: minimal, balanced, or aggressive")
	fs.BoolVar(&flags.optimizeColdStart, "optimize-cold-start", true, "Enable cold start optimization")
	fs.BoolVar(&flags.verbose, "verbose", false, "Enable verbose output")
	fs.StringVar(&flags.signingKey, "signing-key", "", "Ed25519 private key (hex) for signing artifacts")
	fs.DurationVar(&flags.timeout, "timeout", 5*time.Minute, "Compilation timeout")

	help := fs.Bool("help", false, "Show help")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *help {
		printCompileHelp()
		os.Exit(0)
	}

	// Validate required flags
	var missingFlags []string
	if flags.input == "" {
		missingFlags = append(missingFlags, "--input")
	}
	if flags.metadata == "" {
		missingFlags = append(missingFlags, "--metadata")
	}
	if flags.output == "" {
		missingFlags = append(missingFlags, "--output")
	}
	if len(missingFlags) > 0 {
		return nil, fmt.Errorf("missing required flags: %s", strings.Join(missingFlags, ", "))
	}

	// Validate mode
	if flags.mode != ModeDeterministic && flags.mode != ModeComplex && flags.mode != ModeCompatible {
		return nil, fmt.Errorf("invalid mode %q: must be %q, %q, or %q", flags.mode, ModeDeterministic, ModeComplex, ModeCompatible)
	}

	// Validate optimization level
	switch flags.optimize {
	case OptMinimal, OptBalanced, OptAggressive:
		// Valid
	default:
		return nil, fmt.Errorf("invalid optimization level %q: must be %s, %s, or %s",
			flags.optimize, OptMinimal, OptBalanced, OptAggressive)
	}

	return flags, nil
}

// readInputFiles reads and parses the input source and metadata files
func readInputFiles(flags *CompileFlags) (string, *Metadata, error) {
	// Read source file
	sourceCode, err := os.ReadFile(flags.input)
	if err != nil {
		return "", nil, &CompileError{
			Phase:   "input",
			Message: fmt.Sprintf("failed to read input file: %v", err),
			Source:  flags.input,
		}
	}

	// Read metadata file
	metadataContent, err := os.ReadFile(flags.metadata)
	if err != nil {
		return "", nil, &CompileError{
			Phase:   "input",
			Message: fmt.Sprintf("failed to read metadata file: %v", err),
			Source:  flags.metadata,
		}
	}

	// Parse metadata
	var metadata Metadata
	if err := json.Unmarshal(metadataContent, &metadata); err != nil {
		return "", nil, &CompileError{
			Phase:   "input",
			Message: fmt.Sprintf("failed to parse metadata JSON: %v", err),
			Source:  flags.metadata,
		}
	}

	// Validate and set defaults for metadata
	if err := validateMetadata(&metadata); err != nil {
		return "", nil, err
	}

	return string(sourceCode), &metadata, nil
}

// validateMetadata validates metadata and sets defaults
func validateMetadata(metadata *Metadata) error {
	if metadata.Name == "" {
		return &CompileError{
			Phase:   "input",
			Message: "metadata.name is required",
		}
	}

	// Set defaults
	if metadata.Version == "" {
		metadata.Version = "1.0.0"
	}
	if metadata.EntryPoint == "" {
		metadata.EntryPoint = "handler"
	}
	if metadata.Runtime == "" {
		metadata.Runtime = "flypy"
	}

	// Validate name format (alphanumeric, hyphens, underscores)
	for _, r := range metadata.Name {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return &CompileError{
				Phase:   "input",
				Message: fmt.Sprintf("metadata.name contains invalid character: %q", r),
			}
		}
	}

	return nil
}

// validateSourceCode validates the Python source code
func validateSourceCode(source string) error {
	if len(source) == 0 {
		return &CompileError{
			Phase:   "input",
			Message: "source file is empty",
		}
	}

	// 10MB limit
	const maxSourceSize = 10 * 1024 * 1024
	if len(source) > maxSourceSize {
		return &CompileError{
			Phase:   "input",
			Message: fmt.Sprintf("source file exceeds %dMB limit", maxSourceSize/1024/1024),
		}
	}

	return nil
}

// compilePythonToWASM compiles Python source to WASM using the internal/flypy compiler
func compilePythonToWASM(sourceCode string, metadata *Metadata, flags *CompileFlags) (*CompilationResult, error) {
	result := &CompilationResult{
		Success:  true,
		Errors:   []CompileError{},
		Warnings: []string{},
	}

	// Create compiler configuration
	config := &flypy.Config{
		Mode:      flypy.ExecutionMode(flags.mode),
		OutputDir: flags.output,
		Verbose:   flags.verbose,
		Version:   metadata.Version,
	}

	// Parse signing key if provided
	if flags.signingKey != "" {
		keyBytes, err := hex.DecodeString(flags.signingKey)
		if err != nil {
			return nil, &CompileError{
				Phase:   "input",
				Message: fmt.Sprintf("invalid signing key (must be hex): %v", err),
			}
		}
		if len(keyBytes) != ed25519.PrivateKeySize {
			return nil, &CompileError{
				Phase:   "input",
				Message: fmt.Sprintf("invalid signing key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(keyBytes)),
			}
		}
		config.SignKey = keyBytes
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
	defer cancel()

	// Create compiler
	compiler := flypy.NewCompiler(config)

	// Compile
	flypyResult, err := compiler.Compile(ctx, sourceCode, metadata.Name)
	if err != nil {
		result.Success = false

		// Parse error to determine phase
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "parse"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "parse",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "restriction"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "restriction",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "IR") || strings.Contains(errMsg, "ir"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "ir",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "determinism") || strings.Contains(errMsg, "verify"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "verify",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "side effect"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "verify",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "rust") || strings.Contains(errMsg, "Rust"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "codegen",
				Message: errMsg,
			})
		case strings.Contains(errMsg, "wasm") || strings.Contains(errMsg, "WASM") || strings.Contains(errMsg, "compile"):
			result.Errors = append(result.Errors, CompileError{
				Phase:   "compile",
				Message: errMsg,
			})
		default:
			result.Errors = append(result.Errors, CompileError{
				Phase:   "unknown",
				Message: errMsg,
			})
		}

		return result, nil
	}

	// Extract results
	result.WASMBytes = flypyResult.Artifact.WasmModule
	result.Warnings = flypyResult.Warnings
	result.SideEffects = flypyResult.SideEffects
	result.SideEffectStats = flypyResult.SideEffectSummary

	// Generate manifest
	result.Manifest = generateManifest(sourceCode, flypyResult, metadata, flags)

	return result, nil
}

// generateManifest creates the compilation manifest
func generateManifest(sourceCode string, flypyResult *flypy.Result, metadata *Metadata, flags *CompileFlags) map[string]interface{} {
	artifact := flypyResult.Artifact

	manifest := map[string]interface{}{
		"name":                 metadata.Name,
		"version":              metadata.Version,
		"runtime":              metadata.Runtime,
		"entry_point":          metadata.EntryPoint,
		"description":          metadata.Description,
		"mode":                 flags.mode,
		"build_time":           time.Now().Format(time.RFC3339),
		"optimization_level":   flags.optimize,
		"cold_start_optimized": flags.optimizeColdStart,
		"wasm_file":            "function.wasm",
		"wasm_size_bytes":      len(artifact.WasmModule),
		"source_size_bytes":    len(sourceCode),
		"compiler_version":     flypy.GetVersion(),
		"hashes": map[string]string{
			"wasm_sha256":      artifact.DeterminismHash,
			"determinism_hash": artifact.DeterminismHash,
		},
		"capabilities": map[string]interface{}{
			"requested":  artifact.CapabilityMap.Requested,
			"approved":   artifact.CapabilityMap.Approved,
			"restricted": true,
		},
	}

	// Add side effect summary
	if len(flypyResult.SideEffectSummary) > 0 {
		manifest["side_effects"] = map[string]interface{}{
			"summary": flypyResult.SideEffectSummary,
			"pure":    len(flypyResult.SideEffects) == 0,
		}
	}

	// Add determinism proof
	if flypyResult.DeterminismProof != nil {
		manifest["determinism_proof"] = map[string]interface{}{
			"ir_hash":      flypyResult.DeterminismProof.IRHash,
			"timestamp":    flypyResult.DeterminismProof.Timestamp,
			"capabilities": flypyResult.DeterminismProof.Capabilities,
		}
	}

	return manifest
}

// writeOutputFiles writes all compilation outputs to the output directory
func writeOutputFiles(flags *CompileFlags, result *CompilationResult, metadata Metadata) error {
	// Write main WASM file
	wasmPath := filepath.Join(flags.output, "function.wasm")
	if err := os.WriteFile(wasmPath, result.WASMBytes, 0644); err != nil {
		return fmt.Errorf("failed to write WASM file: %w", err)
	}

	// Write state_transition.wasm (compatibility copy)
	stateWasmPath := filepath.Join(flags.output, "state_transition.wasm")
	if err := os.WriteFile(stateWasmPath, result.WASMBytes, 0644); err != nil {
		return fmt.Errorf("failed to write state transition WASM: %w", err)
	}

	// Write manifest
	manifestPath := filepath.Join(flags.output, "manifest.json")
	manifestJSON, err := json.MarshalIndent(result.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Write capabilities
	capabilitiesPath := filepath.Join(flags.output, "capabilities.json")
	capabilitiesJSON, err := json.MarshalIndent(result.Manifest["capabilities"], "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}
	if err := os.WriteFile(capabilitiesPath, capabilitiesJSON, 0644); err != nil {
		return fmt.Errorf("failed to write capabilities: %w", err)
	}

	// Write determinism hash
	hashes, ok := result.Manifest["hashes"].(map[string]string)
	if ok {
		hashPath := filepath.Join(flags.output, "determinism.hash")
		if err := os.WriteFile(hashPath, []byte(hashes["determinism_hash"]), 0644); err != nil {
			return fmt.Errorf("failed to write determinism hash: %w", err)
		}
	}

	// Write side effects report if any
	if len(result.SideEffects) > 0 {
		sideEffectsPath := filepath.Join(flags.output, "side_effects.json")
		sideEffectsJSON, err := json.MarshalIndent(result.SideEffects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal side effects: %w", err)
		}
		if err := os.WriteFile(sideEffectsPath, sideEffectsJSON, 0644); err != nil {
			return fmt.Errorf("failed to write side effects: %w", err)
		}
	}

	return nil
}

// printCompilationHeader prints verbose compilation header
func printCompilationHeader(flags *CompileFlags, metadata *Metadata, sourceLen int) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              FlyPy Python-to-WASM Compiler                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("  Input:       %s\n", flags.input)
	fmt.Printf("  Metadata:    %s\n", flags.metadata)
	fmt.Printf("  Output:      %s\n", flags.output)
	fmt.Printf("  Function:    %s@%s\n", metadata.Name, metadata.Version)
	fmt.Printf("  Entry Point: %s\n", metadata.EntryPoint)
	fmt.Printf("  Mode:        %s\n", flags.mode)
	fmt.Printf("  Optimize:    %s\n", flags.optimize)
	fmt.Printf("  Source Size: %d bytes\n", sourceLen)
	fmt.Println()
	fmt.Println("  Compilation Pipeline:")
}

// printCompilationErrors prints compilation errors
func printCompilationErrors(result *CompilationResult) {
	fmt.Fprintf(os.Stderr, "\n❌ Compilation failed\n\n")

	if len(result.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "Errors:")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  • %s\n", e.Error())
		}
	}
}

// printCompilationSuccess prints success message
func printCompilationSuccess(result *CompilationResult, metadata Metadata, verbose bool) {
	if verbose {
		fmt.Println()
		fmt.Println("  Output Files:")
		fmt.Printf("    • function.wasm (%d bytes)\n", len(result.WASMBytes))
		fmt.Printf("    • manifest.json\n")
		fmt.Printf("    • capabilities.json\n")
		if len(result.SideEffects) > 0 {
			fmt.Printf("    • side_effects.json\n")
		}

		if len(result.Warnings) > 0 {
			fmt.Println()
			fmt.Println("  Warnings:")
			for _, w := range result.Warnings {
				fmt.Printf("    ⚠ %s\n", w)
			}
		}

		fmt.Println()
		fmt.Println("════════════════════════════════════════════════════════════")
	}

	fmt.Printf("✅ Successfully compiled %s to WebAssembly\n", metadata.Name)
}

// isAlphaNumeric checks if a rune is alphanumeric
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func printCompileHelp() {
	fmt.Println("FlyPy Compiler - Compile Python to WebAssembly")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  flypy-go compile [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  --input string              Input Python source file (required)")
	fmt.Println("  --metadata string           Input metadata JSON file (required)")
	fmt.Println("  --output string             Output directory (required)")
	fmt.Println("  --mode string               Execution mode (default \"deterministic\")")
	fmt.Println("                              - deterministic: Full restriction enforcement, pure functions only")
	fmt.Println("                              - complex: Extended stdlib (csv, io, re, datetime) with determinism")
	fmt.Println("                              - compatible: Relaxed restrictions with warnings")
	fmt.Println("  --optimize string           Optimization level (default \"balanced\")")
	fmt.Println("                              - minimal: Fast compilation, larger output")
	fmt.Println("                              - balanced: Good speed/size tradeoff")
	fmt.Println("                              - aggressive: Slow compilation, smallest output")
	fmt.Println("  --optimize-cold-start       Enable cold start optimization (default true)")
	fmt.Println("  --signing-key string        Ed25519 private key (hex) for signing artifacts")
	fmt.Println("  --timeout duration          Compilation timeout (default 5m)")
	fmt.Println("  --verbose                   Enable verbose output")
	fmt.Println("  --help                      Show this help message")
	fmt.Println()
	fmt.Println("METADATA FORMAT (JSON):")
	fmt.Println("  {")
	fmt.Println("    \"name\": \"function-name\",")
	fmt.Println("    \"version\": \"1.0.0\",")
	fmt.Println("    \"entry_point\": \"handler\",")
	fmt.Println("    \"description\": \"Function description\"")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic compilation")
	fmt.Println("  flypy-go compile --input handler.py --metadata meta.json --output ./dist")
	fmt.Println()
	fmt.Println("  # Verbose output with aggressive optimization")
	fmt.Println("  flypy-go compile -i handler.py -m meta.json -o ./dist --optimize aggressive --verbose")
	fmt.Println()
	fmt.Println("  # Compatible mode for migration")
	fmt.Println("  flypy-go compile -i handler.py -m meta.json -o ./dist --mode compatible")
	fmt.Println()
	fmt.Println("RESTRICTIONS:")
	fmt.Println("  The following are not allowed in deterministic mode:")
	fmt.Println("  • eval(), exec(), compile() - dynamic code execution")
	fmt.Println("  • __import__(), importlib - dynamic imports")
	fmt.Println("  • open(), file I/O - use capability system instead")
	fmt.Println("  • os.*, sys.* - system access")
	fmt.Println("  • subprocess, threading, multiprocessing - concurrency")
	fmt.Println("  • socket, pickle, marshal - network and serialization")
	fmt.Println()
	fmt.Println("ALLOWED MODULES:")
	fmt.Println("  json, re, math, random (seeded), datetime, collections,")
	fmt.Println("  functools, operator, itertools, base64, hashlib,")
	fmt.Println("  urllib.parse, html, textwrap, string, unicodedata")
	fmt.Println()
	fmt.Println("REQUIREMENTS:")
	fmt.Println("  • Python 3.x - for AST parsing")
	fmt.Println("  • Rust toolchain - for WASM compilation")
	fmt.Println("  • wasm-pack or cargo - for building WASM")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/functionfly/functionfly")
}

// generateSigningKey generates a new Ed25519 signing key (helper for users)
func generateSigningKey() (ed25519.PrivateKey, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	return privateKey, err
}
