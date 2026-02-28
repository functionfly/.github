# Introducing State Fabric: Composable Durable State for Stateless Functions

**State Fabric** is FunctionFly's composable durable state layer for stateless functions. It gives your serverless workloads a reliable, globally addressable memory layer—so agents, workflows, and functions can share state without losing it.

---

## Why State Fabric?

Today's functions are **stateless by design**. That's great for scale, but it makes coordination, agent memory, and multi-step workflows hard. State Fabric adds a first-class state layer: stores, snapshots, event logs, and triggers—all addressable via simple URIs like `state://tenant/cart` or `memory://tenant/agent-1`.

| Without State Fabric     | With State Fabric                    |
|-------------------------|--------------------------------------|
| Ephemeral execution     | Durable, replayable state            |
| No shared memory         | Globally addressable stores          |
| Ad-hoc coordination     | Event logs and triggers              |
| Stateless-only workflows | Multi-step, agent-friendly workflows |

---

## What You Get

### Durable storage

- **PostgreSQL + pgvector** as the source of truth for structured state and vector embeddings.
- **Optional Redis** for hot cache and low-latency reads when you need them.
- **Deterministic replay** so you can re-run workflows from event history and get the same result.
- **Snapshot management** for point-in-time state and fast recovery.

State is bound to your **fx:// function identity**, so you get sticky infrastructure and storage-based value on top of usage-based compute.

### URI-addressable state

Everything in State Fabric is addressable by a simple URI:

- `state://tenant/store-name` — durable key-value or document store
- `memory://tenant/agent-id` — working or long-term memory for an agent
- `events://tenant/stream` — append-only event log

Example usage in code:

```typescript
// Resolve a store by URI
const cart = await stateFabric.resolve('state://acme/cart');
await cart.merge({ items: [...existing, newItem] });

// Read and write agent memory
const memory = await stateFabric.resolve('memory://acme/support-agent-1');
await memory.append({ role: 'user', content: message });
const context = await memory.recent(20);
```

```go
// Go SDK example
store, err := fabric.Resolve(ctx, "state://acme/inventory")
if err != nil {
    return err
}
defer store.Close()
return store.Set(ctx, "sku-123", inventoryItem)
```

---

## Built for AI Agents

State Fabric is built with **AI agents** in mind:

| Capability        | Description                                              |
|------------------|----------------------------------------------------------|
| **Working memory** | Short-lived context for the current turn or session.   |
| **Long-term memory** | Persistent facts and history across sessions.         |
| **Semantic search**  | pgvector-powered similarity over embeddings.          |
| **Event sourcing**   | Append-only logs for audit and replay.                |

Use it for:

- **Session state** — user preferences, wizard progress, multi-step forms.
- **Cart state** — shopping carts and checkout flows that survive cold starts.
- **Agent context** — conversation history, tool results, and RAG chunks.
- **Workflow state** — durable orchestration across functions and retries.

---

## Architecture at a Glance

```
┌─────────────────────────────────────────────────────────────┐
│                     Your Functions (fx://)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │ state://  memory://  events://
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      State Fabric API                        │
│  (auth, routing, validation, rate limiting)                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  PostgreSQL   │   │  pgvector     │   │  Redis (opt)   │
│  (source of   │   │  (embeddings,  │   │  (hot cache)   │
│   truth)      │   │   similarity)   │   │                │
└───────────────┘   └───────────────┘   └───────────────┘
```

- **Single source of truth** in Postgres; optional Redis for performance.
- **Snapshots** for fast bootstrap; **event log** for replay and audit.
- **Triggers** (coming soon) to react to state changes and drive workflows.

---

## Key Concepts

### 1. Stores

A **store** is a named, durable container for state. You can think of it as a key-value store or a document store, depending on the adapter.

- **Create** a store via the API or dashboard.
- **Bind** it to a function or tenant.
- **Read/write** via the State Fabric API or SDK.

### 2. Snapshots

A **snapshot** is a point-in-time copy of a store’s state. Use snapshots to:

- Restore state after a failure.
- Branch state for testing or what-if flows.
- Comply with retention and audit policies.

### 3. Event log

The **event log** is an append-only stream of events for a store. Every mutation can emit an event; replaying the log reproduces the current state (event sourcing).

- **Deterministic replay** — same events ⇒ same state.
- **Audit trail** — who changed what, and when.
- **Integration** — feed events into queues or other systems.

### 4. Triggers (coming soon)

**Triggers** will let you react to state changes (e.g. “when cart is non-empty, call payment service”). This ties State Fabric into your event-driven and workflow layers.

---

## Getting Started

1. **Open State Fabric** in the FunctionFly dashboard and create your first store.
2. **Bind the store** to a function or use the HTTP API from any runtime.
3. **Read the architecture doc** in the repo for schema, APIs, and limits.
4. **Try the SDK** — we provide TypeScript/JavaScript and Go clients; more runtimes are on the roadmap.

We’re shipping more features—**replay**, **triggers**, and tighter **registry** integration—over the next releases. If you have use cases or feedback, reach out via the dashboard or our docs.

---

## Summary

| Feature            | Status   |
|--------------------|----------|
| Durable stores      | Available |
| Snapshots          | Available |
| Event log          | Available |
| pgvector / search  | Available |
| Redis cache        | Optional  |
| Triggers           | Coming soon |
| Replay UI          | Coming soon |

State Fabric gives your stateless functions a **composable, durable state layer**—so you can build agents, workflows, and shared state without giving up the benefits of serverless.
