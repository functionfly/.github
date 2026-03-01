# DRE Crypto - Merkle Execution Graph (MEG) Primitives

This package implements the cryptographic primitives for the Deterministic Replay Engine (DRE) 2.0 protocol, specifically the ExecutionRootHash v1.0 specification.

## Hash Algorithm: BLAKE3

**BLAKE3** is the default hash algorithm for all MEG computations.

### Why BLAKE3?

- **Extremely fast** - Optimized for modern CPUs
- **Tree-native** - Built-in tree hashing mode for parallel processing
- **Parallelizable** - Can hash multiple chunks simultaneously  
- **Strong cryptographic guarantees** - 256-bit security level
- **Ideal for large trace hashing** - Efficiently handles large execution traces

### Output

- Hash output: 32 bytes (256 bits)
- Hex-encoded output: 64 hex characters

## Protocol Components

### Domain Separation Tags

All hashes use domain separation to prevent cross-layer collision attacks:

| Tag | Purpose |
|-----|---------|
| `FX_INPUT` | Input payload hash |
| `FX_ENV` | Environment/capsule descriptor hash |
| `FX_DEPS` | Dependency Merkle root hash |
| `FX_DEP_NODE` | Individual dependency node hash |
| `FX_TRACE` | Execution trace Merkle root hash |
| `FX_TRACE_CHUNK` | Individual trace chunk hash |
| `FX_RES` | Resource usage hash |
| `FX_OUT` | Output payload hash |
| `FX_META` | Metadata hash |
| `FX_NODE` | Merkle tree node hash |
| `FX_CERT` | Certificate hash |
| `FX_REPLAY_PROOF` | Replay proof hash |

### Merkle Execution Graph (MEG)

The MEG combines 7 component hashes in fixed order to produce the ExecutionRootHash:

1. **InputHash** - Canonical hash of input payload
2. **EnvironmentHash** - Canonical hash of capsule environment
3. **DependencyHash** - Merkle root of dependency tree
4. **TraceHash** - Merkle root of execution trace chunks
5. **ResourceHash** - Canonical hash of resource usage
6. **OutputHash** - Canonical hash of output payload
7. **MetadataHash** - Canonical hash of execution metadata

The final ExecutionRootHash is the Merkle root of these 7 component hashes.

## Key Properties

- ✅ **Deterministic** - Same inputs always produce same hash
- ✅ **Composable** - Component hashes can be verified independently
- ✅ **Partially verifiable** - Can verify subset of components
- ✅ **Replay forgery resistant** - Includes nonce and deterministic seeds
- ✅ **Efficient to compute** - BLAKE3 is highly optimized
- ✅ **Blockchain-anchor friendly** - Single 32-byte root hash

## Usage

```go
package main

import (
    "fmt"
    drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

func main() {
    // Build a MEG
    components := drecrypto.MEGComponents{
        InputPayload: map[string]interface{}{
            "args": []string{"arg1", "arg2"},
            "caller": "fx://example/func@1.0.0",
        },
        EnvironmentData: map[string]string{
            "runtime": "wasm/1.0",
        },
        Dependencies: []drecrypto.Dependency{
            {Name: "lodash", Version: "4.17.21", ContentHash: "abc123..."},
        },
        ResourceUsage: map[string]int64{
            "cpu_cycles": 1000000,
            "memory_peak": 1024,
        },
        OutputPayload: map[string]interface{}{
            "result": "success",
        },
        Metadata: map[string]string{
            "execution_id": "exec_123",
            "function_id": "fx://example/func",
        },
    }
    
    result, err := drecrypto.BuildMEG(components)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("ExecutionRootHash: %s\n", result.ExecutionRootHash)
}
```

## Migration from SHA-256

This package previously used SHA-256 as the default hash algorithm. To maintain backward compatibility with legacy systems, you can build with the `sha256` build tag:

```bash
go build -tags sha256 .
```

However, BLAKE3 is recommended for all new deployments.

## Files

- `hasher.go` - Hash functions and Merkle root computation
- `canonicalize.go` - RFC-8785 JSON canonicalization
- `meg.go` - Merkle Execution Graph builder
