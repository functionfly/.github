# FlyMind AI Service - Production Deployment Guide

This document covers deploying the AI Service to Fly.io for production use.

## Architecture

```
┌─────────────────────┐         ┌─────────────────────┐
│   functionfly-api   │         │ functionfly-ai-     │
│   (Orchestrator)    │◄───────►│ service             │
│   Port: 8080        │  .internal   Port: 8081      │
│   Public: Yes       │  (private)   Public: No       │
└─────────────────────┘         └─────────────────────┘
```

The AI Service runs on Fly.io with **internal networking only** - no public exposure. The orchestrator communicates via Fly's private `.internal` DNS.

## Prerequisites

1. **Fly CLI installed**: `curl -L https://fly.io/install.sh | sh`
2. **Logged in**: `fly auth login`
3. **Same organization** as the orchestrator app (functionfly-api)

## Quick Start

### 1. Create the Fly.io App

```bash
cd ai-service

# Create app (only needed once)
fly apps create functionfly-ai-service
```

### 2. Set Required Secrets

```bash
# Database (same Neon PostgreSQL as orchestrator)
fly secrets set DATABASE_URL="postgresql://..." \
              --app functionfly-ai-service

# Redis (same Redis/Upstash as orchestrator)
fly secrets set REDIS_ADDR="your-redis-host:6379" \
              REDIS_PASSWORD="your-redis-password" \
              --app functionfly-ai-service

# LLM Provider API Keys
fly secrets set OPENAI_API_KEY="sk-..." \
              ANTHROPIC_API_KEY="sk-ant-..." \
              --app functionfly-ai-service

# Optional: OpenRouter for model routing
fly secrets set OPENROUTER_API_KEY="sk-or-..." \
              --app functionfly-ai-service

# Orchestrator API Key (for callbacks)
fly secrets set ORCHESTRATOR_API_KEY="your-secure-random-key" \
              --app functionfly-ai-service
```

### 3. Deploy

```bash
# Using the deploy script
./deploy.sh

# Or manually
fly deploy --app functionfly-ai-service
```

### 4. Update Orchestrator

The orchestrator needs to know where to find the AI service:

```bash
fly secrets set AI_SERVICE_URL="http://functionfly-ai-service.internal:8081" \
              AI_SERVICE_API_KEY="same-key-as-above" \
              --app functionfly-api
```

## Verification

### From the Orchestrator

```bash
# SSH into orchestrator
fly ssh console --app functionfly-api

# Test AI service health
curl http://functionfly-ai-service.internal:8081/health
```

### From Your Local Machine (Temporary)

```bash
# Start a proxy
fly proxy 8081:8081 --app functionfly-ai-service

# In another terminal
curl http://localhost:8081/health

# View docs (if temporarily exposed)
curl http://localhost:8081/docs
```

## Monitoring

```bash
# View logs
fly logs --app functionfly-ai-service

# Check status
fly status --app functionfly-ai-service

# Check machine list
fly machines list --app functionfly-ai-service
```

## Troubleshooting

### Health Check Fails

```bash
# Check recent logs
fly logs --app functionfly-ai-service -n 100

# SSH and check directly
fly ssh console --app functionfly-ai-service
curl localhost:8081/health
```

### Orchestrator Can't Connect

Verify the orchestrator can reach the AI service:

```bash
fly ssh console --app functionfly-api

# Should return JSON health response
curl -v http://functionfly-ai-service.internal:8081/health
```

### Secrets Not Set

```bash
# List current secrets
fly secrets list --app functionfly-ai-service

# Set missing ones
fly secrets set KEY=value --app functionfly-ai-service
```

## Scaling

Edit `fly.toml` to adjust resources:

```toml
[[vm]]
  size = "shared-cpu-4x"    # Upgrade to 4x for more CPU
  memory = "2048mb"         # More RAM for heavy embeddings
  cpus = 2
```

Or scale manually:

```bash
# Vertical scaling
fly scale vm shared-cpu-4x --app functionfly-ai-service
fly scale memory 2048 --app functionfly-ai-service

# Horizontal scaling (if you add HTTP service later)
fly scale count 2 --app functionfly-ai-service
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | Neon PostgreSQL connection |
| `REDIS_ADDR` | Yes | - | Redis host:port |
| `REDIS_PASSWORD` | No | - | Redis password |
| `OPENAI_API_KEY` | Yes* | - | OpenAI API key |
| `ANTHROPIC_API_KEY` | Yes* | - | Anthropic API key |
| `ORCHESTRATOR_URL` | No | internal | Orchestrator callback URL |
| `ORCHESTRATOR_API_KEY` | No | - | Auth for orchestrator |

*At least one LLM provider required.

## CI/CD

GitHub Actions workflow included at `.github/workflows/ai-service-deploy.yml`.

**Required Secret:**
- `FLY_API_TOKEN` - Generate with: `fly tokens create deploy --name github-actions`

## Architecture Notes

### Why Internal Networking?

- **Security**: AI service not exposed to internet
- **Latency**: Fly's internal network is faster than public routes
- **Simplicity**: No need for CORS, TLS termination, or auth at AI service level

### Why Separate App?

- **Independent scaling**: AI workloads (embeddings) differ from API traffic
- **Independent deployment**: Update AI service without touching orchestrator
- **Resource isolation**: CPU/memory intensive AI ops don't affect API latency
