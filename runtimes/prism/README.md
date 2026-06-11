# FunctionFly Prism Runtime

## Universal Adaptive WASM Execution Fabric

> "Write once. Execute anywhere. Coordinate everywhere."

---

## Overview

FunctionFly Prism is a distributed runtime for AI-native execution, autonomous agents, robotics, edge systems, and cross-language functions. Prism treats every execution unit as a self-describing, portable, AI-aware component within an intelligent execution fabric.

### Key Differentiators

- **Adaptive Execution Cells (AECs)** — Self-describing, portable WASM cells with lifecycle management, migration support, and status tracking
- **Universal Capability Discovery** — "The DNS of AI capabilities" — robots, AI agents, browsers, drones, SaaS apps, and IDEs can all discover compatible functions dynamically
- **StateStream Memory Fabric** — Distributed streaming memory with CRDT synchronization and event sourcing
- **Quantum Snapshotting** — Cell freeze, state serialization, and migration support
- **Neural Execution Optimization** — Q-learning based optimizer with reinforcement learning structure for self-optimization

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    FunctionFly Mesh                         │
│  ┌───────────────────────────────────────────────────────┐  │
│  │         Global Function Coordination Layer             │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                  │
│         ┌─────────────────┼─────────────────┐                │
│         ▼                 ▼                 ▼                │
│  ┌───────────┐     ┌───────────┐     ┌───────────┐          │
│  │Cloud Node │     │Edge Node  │     │Browser    │          │
│  └───────────┘     └───────────┘     └───────────┘          │
│         │                 │                 │                │
│         └─────────────────┼─────────────────┘                │
│                           ▼                                  │
│  ┌───────────────────────────────────────────────────────┐   │
│  │      Adaptive Execution Cells (WASM)                   │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐               │   │
│  │  │ Cell A  │  │ Cell B  │  │ Cell C  │  ...          │   │
│  │  └─────────┘  └─────────┘  └─────────┘               │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## Core Runtime Layers

### 1. HyperCore Scheduler

An AI-aware distributed scheduler that decides:

- **Where** functions execute (cloud, edge, browser, robotic, mobile, IoT)
- **GPU vs CPU** placement based on workload characteristics
- **Memory optimization** through predictive allocation
- **Latency-aware** placement using placement hints
- **Cost-aware** execution for multi-tenant optimization
- **AI-model affinity** routing for inference workloads

**Key Features:**
- Multi-tenant isolation with resource guarantees
- Placement scoring based on latency, cost, GPU affinity, and availability
- Configurable concurrent execution limit (default: 64)
- Node registration limiting based on max queue size

### 2. WASM Fusion Engine

Dynamic execution graphs with WASM module composition:

- **Streaming function compositions** — Chain AI pipelines without separate infrastructure
- **Directed graph execution** — Multiple WASM modules executed in topological order
- **Optional runtime patching** — Live patching support (disabled by default)
- **WASI preview1 and preview2 support** — Standard WASI interfaces

**Example Pipeline:**
```text
Speech Input → Transcription WASM → Reasoning Agent → Action Planner → Robot Control
```

### 3. Universal Capability Layer (UCL)

Everything becomes a discoverable capability:

```json
{
  "capability": "vision.detect",
  "latency": "12ms",
  "trust": 0.998,
  "gpu_required": true
}
```

This enables:
- Robots discovering manipulation capabilities
- AI agents finding reasoning services
- Browsers locating compute resources
- Drones accessing sensor fusion
- SaaS apps composing functions dynamically

**Think: "The DNS of AI capabilities."**

### 4. StateStream Memory Fabric

Distributed streaming memory with:

- Event-sourced state for audit and replay
- Resumable execution after interruption
- Memory snapshots for migration
- CRDT synchronization for distributed consistency (LWW, GCounter, PnCounter)
- Offline reconciliation
- Redis-backed persistence (optional)

**Enables:**
- Long-running AI agents with persistent memory
- Robotic systems with fault-tolerant workflows
- Collaborative multi-agent swarms
- Edge computing with intermittent connectivity

### 5. Quantum Snapshotting

Cell state persistence and migration:

- **Freeze execution** — Pause cell at any checkpoint
- **Serialize runtime state** — Memory, CPU, open handles
- **Migration support** — Move cells between nodes
- **Compression** — zstd and lz4 compression for snapshots

**Use Cases:**
- **Failover** — Move from a failing node to a healthy one
- **Cost optimization** — Migrate based on load
- **Mobile robotics** — Handoff cognition between local and cloud nodes
- **Edge handoff** — Seamlessly transfer execution as devices move

### 6. Mesh Networking

P2P capability mesh using libp2p:

- Peer-to-peer communication without central brokers
- Capability discovery through Kademlia DHT
- Local peer discovery via mDNS
- NAT traversal via relay
- State synchronization via CRDT-based protocols
- TCP transport with noise encryption and yamux multiplexing

### 7. Neural Execution Optimization

The runtime has Q-learning based optimization:

- Optimal execution path suggestions
- Hot function identification (cache-worthy)
- Memory behavior profiling
- GPU allocation strategy hints
- Agent coordination pattern learning

**Implementation:**
- Q-learning structure exists for optimization decisions
- Execution profiles track duration, memory, input size
- Optimization suggestions include memory, timeout, and placement recommendations

### 8. Autonomous Function Swarms

Functions can:

- **Spawn sub-functions** — Decompose complex tasks
- **Delegate workloads** — Distribute based on capability
- **Form temporary clusters** — Collaborate on shared goals
- **Self-heal** — Detect and replace failed cells via health monitoring

---

## Adaptive Execution Cells (AECs)

Every execution unit is a self-describing, portable, AI-aware WASM cell.

**Structure:**
```
┌─────────────────────────────────────────────────┐
│                 Execution Cell                    │
├─────────────────────────────────────────────────┤
│  ID: uuid-xxxx-xxxx                             │
│  Status: Running                                 │
│  Config: { memory: 128MB, timeout: 30s }         │
│  Resources: { vcpus: 2, gpu: true }             │
│  Metadata: { name: vision-detector, v1.2.0 }    │
│                                                  │
│  State Slices: [ slice-1, slice-2, ... ]         │
│  Checkpoint Epoch: 42                            │
└─────────────────────────────────────────────────┘
```

---

## Universal WASM Package (.ffpkg)

A standard package format for Prism:

```json
{
  "manifest_version": "1.0.0",
  "metadata": {
    "name": "vision.detect",
    "version": "1.0.0",
    "runtime": "wasm",
    "languages": ["rust", "python"],
    "capabilities": ["gpu", "inference"]
  },
  "modules": [
    {
      "module_id": "mod-001",
      "name": "detector",
      "language": "rust",
      "bytecode": "<binary>",
      "entry_point": "detect"
    }
  ],
  "resources": [],
  "signature": {
    "algorithm": "ed25519",
    "signature": "<base64>"
  }
}
```

---

## Security

**Zero-Trust Execution:**
- Every function can be cryptographically signed (Ed25519)
- WASM bytecode validation via wasmparser
- Execution scope-limiting via capability permissions
- Policies constrain resource access

**OS-Level Enforcement (Linux):**
- **seccomp** — Syscall allowlisting via libseccomp
- **landlock** — Filesystem access restrictions (Linux 5.13+)
- **Resource limits** — Memory, CPU, file descriptors, processes

**Enclave Detection:**
- SGX, SEV, and TrustZone detection support
- Actual TEE runtime integration is detection-only

**Not Yet Implemented:**
- Post-quantum signatures (roadmap item)
- Deterministic audit logs (standard logging present)

---

## Developer Experience

### CLI

```bash
# Start runtime
prism start --address 0.0.0.0:8080 --mesh

# Cell management
prism cell create --module detector.wasm --memory 128
prism cell list
prism cell terminate <cell_id>
prism cell snapshot <cell_id>
prism cell migrate <cell_id> --target <node>

# Capability discovery
prism capability register --name vision.detect --category AI
prism capability discover "image classification"
prism capability list

# Swarm operations
prism swarm create my-swarm
prism swarm join my-swarm
prism swarm leave my-swarm
prism swarm list
prism swarm command my-swarm --cmd coordinate

# Package management
prism package build --source ./src --output my-package.ffpkg
prism package inspect my-package.ffpkg
prism package sign my-package.ffpkg --key key.pem

# Interactive REPL
prism repl

# Runtime status
prism status

# Generate documentation
prism doc --output ./docs
```

### HTTP API

The runtime exposes a JSON HTTP API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/cells` | Create cell |
| GET | `/cells` | List all cells |
| GET | `/cells/{id}` | Get specific cell |
| POST | `/cells/{id}/snapshot` | Snapshot cell |
| GET | `/cells/{id}/snapshots` | List cell snapshots |
| POST | `/execute` | Execute cell |
| POST | `/snapshots/{id}/restore` | Restore from snapshot |
| DELETE | `/snapshots/{id}` | Delete snapshot |
| POST | `/capabilities` | Register capability |
| POST | `/capabilities/invoke` | Invoke capability |
| GET | `/capabilities` | List capabilities |
| POST | `/swarms` | Create swarm |
| GET | `/swarms` | List swarms |
| POST | `/swarms/{id}/join` | Join swarm |
| POST | `/swarms/{id}/leave` | Leave swarm |
| GET | `/optimize/{cell_id}` | Get optimization suggestion |

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| **Runtime Core** | Rust, Wasmtime 45.0.1, Tokio |
| **WASM Engine** | wasmtime 45.0.1 (WASI preview1/preview2) |
| **Networking** | libp2p 0.54 (TCP + noise + yamux), async-nats client |
| **State** | Custom CRDT (LWW, GCounter, PnCounter), Event sourcing |
| **Storage** | Redis (optional, via state-stream feature) |
| **Serialization** | Protobuf (prost), MsgPack (rmp-serde), CBOR (ciborium), JSON |
| **AI/Optimization** | Q-learning reinforcement learning (custom implementation) |
| **Security** | libseccomp, landlock, ed25519-dalek, wasmparser |
| **Compression** | zstd, lz4 |

---

## Project Structure

```
runtimes/prism/
├── Cargo.toml
├── build.rs
├── proto/
│   └── prism.proto          # Protocol Buffer definitions
├── src/
│   ├── main.rs              # CLI binary and HTTP server
│   ├── lib.rs               # Library root
│   ├── core/                # Core types (cells, errors, metrics)
│   ├── hypercore/           # Scheduler and placement
│   ├── wasm_fusion/         # WASM Fusion Engine
│   ├── ucl/                 # Universal Capability Layer
│   ├── state_stream/        # Distributed memory fabric
│   ├── quantum/             # Snapshotting and migration
│   ├── mesh/                # P2P networking
│   ├── neural/              # RL-based optimization
│   ├── swarm/               # Autonomous coordination
│   ├── cli/                 # CLI and REPL
│   ├── codec/               # Serialization codecs
│   ├── security/            # Security enforcement
│   ├── proto/               # Protobuf generated types
│   ├── runtime.rs           # Runtime context
│   ├── nats_client.rs       # NATS client
│   ├── integration_tests.rs # Integration tests
│   └── benches/             # Benchmark functions
├── benches/                 # Criterion benchmark harness
└── README.md
```

---

## Status

### Core Systems Implemented

| Component | Status | Description |
|-----------|--------|-------------|
| **Adaptive Execution Cells** | ✅ | Cell lifecycle, migration, status tracking |
| **HyperCore Scheduler** | ✅ | AI-aware placement, node management, scoring |
| **WASM Fusion Engine** | ✅ | wasmtime execution, graph composition |
| **Universal Capability Layer** | ✅ | Registry, discovery, trust scoring, matching |
| **StateStream Memory** | ✅ | CRDT sync, event sourcing, Redis persistence |
| **Quantum Snapshotting** | ✅ | Checkpoint/restore, compression, migration |
| **Mesh Networking** | ✅ | libp2p P2P, DHT discovery, mDNS, relay |
| **Neural Optimization** | ✅ | Q-learning structure, feedback loop, profiling |
| **Autonomous Swarms** | ✅ | Coordinator, self-healing, swarm commands |
| **NATS Integration** | ✅ | Orchestrator communication |
| **HTTP API Server** | ✅ | REST endpoints for all runtime operations |
| **Security Enforcement** | ✅ | seccomp, landlock, WASM validation |
| **Integration Tests** | ✅ | **26 integration tests** |
| **Performance Benchmarks** | ✅ | **58+ benchmark functions** |

### Running Tests

```bash
# Run all tests
cd runtimes/prism && cargo test

# Run with output
cargo test -- --nocapture

# Run integration tests only
cargo test integration_tests
```

### Running Benchmarks

```bash
# Run all benchmarks
cd runtimes/prism && cargo bench

# Run specific benchmark group
cargo bench -- cell_creation

# Generate HTML report
cargo bench -- --profile-time
```

Benchmarks cover:
- **Cell benchmarks**: Creation, ID generation, status transitions, migration eligibility, cloning
- **Scheduler benchmarks**: Node registration, scheduling decisions, stats collection
- **CRDT benchmarks**: LWW, GCounter, PnCounter operations, merging
- **Neural benchmarks**: Q-learning updates, suggestion generation, policy management
- **WASM benchmarks**: Module compilation, registration, fusion graph operations

---

## Contributing

This is a ground-up implementation. See the FunctionFly contributing guidelines for details.

**Design Philosophy:**
- Keep code DRY — leverage existing patterns in `runtimes/local` and [functionfly/sar](https://github.com/functionfly/sar)
- Use protobuf for service definitions
- Design for scale from day one
- Prefer composable services over monoliths
