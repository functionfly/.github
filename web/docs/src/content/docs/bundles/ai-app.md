---
title: AI App Bundle
description: Ship AI-powered applications faster with LLM integrations, Vector DB, Prompt management, and Usage tracking.
sidebar:
  order: 4
---

# AI App Bundle

The complete backend for AI-powered applications. Everything you need to build, deploy, and scale AI features.

## What's Included

### LLM Integrations
- **Multiple providers** — OpenAI, Anthropic, Google, Azure, self-hosted
- **Model routing** — Automatic failover and cost optimization
- **Caching layer** — Reduce costs with semantic caching
- **Rate limiting** — Per-user and per-model limits

### Vector Database
- **Embedding storage** — Store and search embeddings at scale
- **Semantic search** — Natural language queries over your data
- **Metadata filtering** — Filter by date, category, tags
- **Hybrid search** — Combine vector and keyword search

### Prompt Management
- **Version control** — Track prompt changes over time
- **A/B testing** — Compare prompt versions
- **Variables & templates** — Dynamic prompt interpolation
- **Chain support** — Multi-step reasoning pipelines

### Usage Tracking
- **Per-user metrics** — Track usage by user and API key
- **Cost attribution** — Assign costs to products, features
- **Rate alerts** — Notify when usage exceeds thresholds
- **Budget controls** — Hard limits per user or project

### Memory & Context
- **Short-term memory** — Conversation context
- **Long-term memory** — Persistent user preferences
- **Semantic memory** — Learned facts and patterns
- **Episodic memory** — Past interactions and outcomes

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Your Frontend                          │
│              (Chat, Agents, RAG, etc.)                       │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                     AI App Bundle                           │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │   LLM   │  │ Vector  │  │ Prompt  │  │  Usage  │        │
│  │ Gateway │  │   DB    │  │ Manager │  │ Tracker │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
│  ┌────▼────────────▼────────────▼────────────▼────┐        │
│  │         Database (PostgreSQL + pgvector)         │       │
│  │ Embeddings │ Prompts │ Usage │ Memory │ Sessions │      │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## Pricing

| Plan | Price | AI Calls | Features |
|------|-------|----------|----------|
| **Founder** | Free (3 months) | 100K/mo | All features |
| **Starter** | $39/mo | 500K/mo | All features |
| **Growth** | $119/mo | 2M/mo | + Advanced analytics |
| **Scale** | $349/mo | Unlimited | + Priority support |

**AI call pricing** is provider-pass-through (OpenAI, Anthropic, etc.)

## Getting Started

1. Go to **Dashboard → Bundles → AI App**
2. Click **Deploy Bundle**
3. Connect your LLM provider (OpenAI, Anthropic, etc.)
4. Configure your embedding model
5. Start building your AI features

## Use Cases

### RAG (Retrieval-Augmented Generation)
Build AI that knows your data. Ingest documents, store embeddings, and retrieve context at query time.

### AI Agents
Create autonomous agents that can use tools, maintain memory, and execute multi-step tasks.

### Chat Applications
Build conversational AI with context windows, conversation history, and user preferences.

## Customization

- Add custom LLM providers
- Configure embedding models per data type
- Build custom prompt chains
- Extend memory types for your use case

## Next Steps

- [Build your first RAG pipeline](/guides/creating-functions/)
- [Set up agent memory](/agents/memory/)
- [Configure usage tracking](/guides/monitoring/)