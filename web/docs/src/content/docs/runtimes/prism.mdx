---
title: Prism Runtime
description: FunctionFly's next-generation distributed WASM runtime for AI workloads, autonomous agents, and edge execution.
sidebar:
  order: 12
---

import { Card, CardGrid } from '@astrojs/starlight/components';

FunctionFly Prism is a next-generation runtime designed for AI-native execution, autonomous agents, and distributed computing. It goes beyond traditional serverless runtimes to provide adaptive, migratable, and AI-aware function execution.

<CardGrid>
  <Card title="Adaptive Execution" icon="seti:custom">
    Functions that can migrate and scale across cloud, edge, and device environments.
  </Card>
  <Card title="AI-Optimized" icon="brain">
    Runtime that learns and optimizes execution patterns for AI workloads.
  </Card>
  <Card title="Universal Portability" icon="globe">
    Execute the same function anywhere—from cloud to edge to browser.
  </Card>
</CardGrid>

## What is Prism?

Prism is FunctionFly's universal execution fabric built on WebAssembly (WASM). Unlike traditional runtimes that simply execute code, Prism treats every function as an **Adaptive Execution Cell (AEC)**—a self-describing, portable, and migratable execution unit.

### Key Capabilities

| Feature | Description |
|---------|-------------|
| **Live Migration** | Functions can pause, snapshot, and resume on different machines |
| **AI-Aware Scheduling** | Intelligent placement based on latency, cost, and workload characteristics |
| **State Streaming** | Distributed memory with event sourcing and deterministic replay |
| **Capability Discovery** | Functions can discover and compose other capabilities dynamically |
| **Swarm Coordination** | Multiple functions can collaborate as autonomous swarms |

## When to Use Prism

Prism is ideal for:

- **Long-running AI agents** that need persistent memory across executions
- **Robotic systems** requiring fault-tolerant workflows with state recovery
- **Edge computing** where functions must migrate between devices
- **Distributed AI pipelines** that compose multiple models dynamically
- **Collaborative multi-agent systems** with shared state requirements

## Supported Languages

| Language | Status | Notes |
|----------|--------|-------|
| **Rust** | Primary | Best performance and safety |
| **Python** | Supported | Via Pyodide compilation |
| **JavaScript/TypeScript** | Supported | Via esbuild bundling |
| **Go** | Supported | Via TinyGo |
| **C/C++** | Supported | Via WASM compilation |

## Function Structure

Prism functions are packaged as `.ffpkg` (FunctionFly Package) files:

```json
{
  "manifest_version": "1.0.0",
  "metadata": {
    "name": "my-function",
    "version": "1.0.0",
    "runtime": "wasm"
  }
}
```

## State Management

### StateStream Memory

Prism provides built-in state streaming with:

- **Event sourcing** — All state changes are logged for audit and replay
- **CRDT synchronization** — eventual consistency across distributed nodes
- **Deterministic replay** — Reproduce any past state
- **Temporal rollback** — Debug by replaying execution history

### Example: Persistent Agent Memory

```python
async def handler(request, context):
    # Access persistent state from previous executions
    memory = context.state.get("agent_memory", {})

    # Update state
    memory["interactions"] = memory.get("interactions", 0) + 1
    context.state.set("agent_memory", memory)

    return {"status": 200, "body": {"interactions": memory["interactions"]}}
```

## Execution Model

```
Request → HyperCore Scheduler → WASM Fusion Engine → Execution Cell
                                              ↓
                                      StateStream Memory
                                              ↓
                                          Response
```

### HyperCore Scheduler

The AI-aware scheduler determines:

- **Where** to execute (cloud, edge, browser, IoT)
- **When** to migrate for cost optimization
- **How** to balance latency vs. throughput

### WASM Fusion Engine

Enables dynamic function composition:

```text
Input → Module A → Module B → Module C → Output
```

Without spawning separate infrastructure for each step.

## Migration and Snapshots

### Quantum Snapshotting

Functions can be frozen and migrated:

1. **Checkpoint** — Snapshot full runtime state
2. **Serialize** — Pack memory, CPU context, open handles
3. **Migrate** — Transfer to target machine
4. **Resume** — Continue execution seamlessly

### Use Cases

- **Cost optimization** — Migrate from expensive cloud to cheap edge during low traffic
- **Failover** — Move from failing node to healthy one
- **Handoff** — Transfer execution as mobile devices change location

## Capabilities and Discovery

### Universal Capability Layer

Prism functions can discover and use capabilities dynamically:

```json
{
  "capability": "vision.detect",
  "latency": "12ms",
  "trust": 0.998
}
```

This enables:
- Robots discovering manipulation services
- AI agents finding reasoning engines
- Browsers locating compute resources
- SaaS apps composing functions at runtime

## Swarm Coordination

Multiple Prism functions can form autonomous swarms:

- **Spawn sub-functions** — Decompose complex tasks
- **Negotiate resources** — Dynamic allocation based on capability
- **Delegate workloads** — Distribute based on specialty
- **Self-heal** — Detect and replace failed cells

## Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| **Timeout** | 300s | 900s |
| **Memory** | 512 MB | 4 GB |
| **State Size** | 100 MB | 1 GB |
| **Snapshot Size** | 50 MB | 500 MB |

## Best Practices

1. **Design for migration** — Keep functions stateless when possible
2. **Use StateStream** — Leverage event sourcing for complex state needs
3. **Set placement hints** — Guide the scheduler with latency/cost preferences
4. **Monitor metrics** — Track execution patterns for optimization
5. **Handle interrupts** — Design for graceful migration mid-execution

## Getting Started

Prism functions are created like any other FunctionFly function:

```bash
# Create a new function (defaults to Prism runtime)
ffly create my-agent --runtime prism

# Deploy
ffly deploy
```

Contact FunctionFly support to enable Prism for your account if you don't see it as an option.

## Related Topics

- [WASM & WebAssembly](/wasm/) — How WASM execution works on FunctionFly
- [Execution](/execution/) — General execution model
- [StateFabric](/guides/statefabric/) — Durable state for functions