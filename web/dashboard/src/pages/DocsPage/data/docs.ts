import {
  Book, Rocket, Zap, Shield, Terminal, Code, Settings,
  Globe, Database, Cpu, Layout, Cloud, Workflow, Layers,
  Key, Lock, Eye, FileText, Gauge, AlertTriangle, CheckCircle,
  GitBranch, Package, Boxes, Server, Network, Clock
} from "lucide-react";
import { type LucideIcon } from "lucide-react";

export interface DocPage {
  slug: string;
  title: string;
  description: string;
  content: string;
  lastUpdated?: string;
}

export interface DocSection {
  id: string;
  title: string;
  icon: LucideIcon;
  pages: DocPage[];
}

export const docSections: DocSection[] = [
  {
    id: "get-started",
    title: "Get Started",
    icon: Rocket,
    pages: [
      {
        slug: "welcome",
        title: "Welcome",
        description: "Introduction to FunctionFly - the multi-provider serverless platform",
        content: `
# Welcome to FunctionFly

FunctionFly is a **multi-provider serverless platform** that lets you deploy functions to multiple edge providers simultaneously. Deploy once, run everywhere.

## What is FunctionFly?

FunctionFly abstracts away the complexity of deploying to different serverless platforms. Instead of managing deployments to Vercel, Cloudflare Workers, Fly.io, and Deno Deploy separately, you write your function once and we handle the rest.

### Key Features

- **Multi-Provider Deployment**: Deploy to Vercel, Cloudflare, Fly.io, and Deno Deploy simultaneously
- **Intelligent Failover**: Automatic traffic routing when providers experience issues
- **Unified API**: Single API for all your serverless functions
- **Edge Optimization**: Deploy to 200+ locations worldwide
- **Real-time Analytics**: Monitor performance across all providers

## Quick Start

Get up and running in under 5 minutes:

\`\`\`bash
# Install the fly CLI (Go 1.25+)
go install github.com/functionfly/functionfly/cmd/fly@latest
# Ensure $(go env GOPATH)/bin is on your PATH

# Login to your account
fly login

# Create a new function
fly init my-function

# Deploy
fly deploy
\`\`\`

## Supported Providers

| Provider | Status | Regions |
|----------|--------|---------|
| Vercel | ✅ Ready | 14+ |
| Cloudflare Workers | ✅ Ready | 300+ |
| Fly.io | ✅ Ready | 30+ |
| Deno Deploy | ✅ Ready | 35+ |

## Next Steps

- [Quickstart Guide](/docs/quickstart) - Deploy your first function
- [Core Concepts](/docs/concepts) - Understand the FunctionFly architecture
- [CLI Overview](/docs/cli-overview) - Install and use the CLI
- [CLI Commands](/docs/cli-commands) - Full command reference
- [API Reference](/docs/api-overview) - REST API and authentication
- [Troubleshooting](/docs/troubleshooting-overview) - Common issues and debugging
        `,
        lastUpdated: "2026-02-26"
      },
      {
        slug: "quickstart",
        title: "Quickstart",
        description: "Deploy your first function in under 5 minutes",
        content: `
# Quickstart Guide

Get your first function deployed to multiple edge providers in minutes.

## Prerequisites

- Go 1.25+ (for \`go install\`) or a \`fly\` binary from [GitHub Releases](https://github.com/functionfly/functionfly/releases)
- A FunctionFly account (sign up at [functionfly.com](https://functionfly.com))

## Step 1: Install the CLI

\`\`\`bash
go install github.com/functionfly/functionfly/cmd/fly@latest
\`\`\`

## Step 2: Authenticate

\`\`\`bash
fly login
\`\`\`

This will open a browser window for authentication.

## Step 3: Create a Function

\`\`\`bash
fly init hello-world
\`\`\`

This creates a new function with the following structure:

\`\`\`
hello-world/
├── functionfly.jsonc    # Function configuration
├── main.py              # Function code
└── README.md
\`\`\`

## Step 4: Deploy

\`\`\`bash
cd hello-world
fly deploy
\`\`\`

Your function will be deployed to all configured providers.

## Step 5: Test

\`\`\`bash
curl https://fx.functionfly.io/your-username/hello-world
\`\`\`

That's it! You've deployed your first multi-provider function.
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "concepts",
        title: "Core Concepts",
        description: "Understanding the FunctionFly architecture",
        content: `
# Core Concepts

Understanding these key concepts will help you get the most out of FunctionFly.

## Functions

A Function is the basic unit of deployment in FunctionFly. It's a piece of code that:
- Responds to HTTP requests
- Runs in a serverless environment
- Can be deployed to multiple providers

## Providers

Providers are the edge platforms where your functions run:

### Vercel Functions
- Great for Next.js integration
- Automatic scaling
- Edge and Node.js runtimes

### Cloudflare Workers
- 300+ edge locations
- V8 isolates for fast cold starts
- Durable Objects for state

### Fly.io
- Container-based deployment
- Persistent volumes available
- Machine-based scaling

### Deno Deploy
- Native TypeScript support
- No build step required
- Edge caching built-in

## Routing

FunctionFly automatically routes traffic between providers:

- **Geographic**: Routes to the nearest edge location
- **Performance**: Routes to the fastest responding provider
- **Health**: Automatically fails over to healthy providers

## State Fabric

State Fabric provides stateful capabilities for serverless functions:

- **KV Store**: Key-value storage with edge caching
- **Pub/Sub**: Real-time messaging between functions
- **Queues**: Reliable message queuing
- **Sessions**: User session management

## Adapters

Adapters normalize requests between different provider formats:

| Feature | Vercel | Cloudflare | Fly.io | Deno |
|---------|--------|------------|--------|------|
| Headers | ✓ | ✓ | ✓ | ✓ |
| Query Params | ✓ | ✓ | ✓ | ✓ |
| Body | ✓ | ✓ | ✓ | ✓ |
| Environment | ✓ | ✓ | ✓ | ✓ |
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "cli",
    title: "CLI",
    icon: Terminal,
    pages: [
      {
        slug: "cli-overview",
        title: "Overview",
        description: "Install and configure the FunctionFly CLI",
        content: `
# CLI Overview

The FunctionFly CLI (\`fly\`) is the official tool to create, run, and publish functions from your terminal. Go from idea to global API in under 60 seconds.

## Installation

### Go install (recommended)

\`\`\`bash
go install github.com/functionfly/functionfly/cmd/fly@latest
\`\`\`

Ensure \`$(go env GOPATH)/bin\` is on your \`PATH\`.

### GitHub Releases

Download a prebuilt binary for your OS/arch from the [releases page](https://github.com/functionfly/functionfly/releases) and add it to your \`PATH\`.

### Homebrew (when a tap is published)

\`\`\`bash
brew install functionfly/tap/ffly
\`\`\`

If the tap is not available yet, use Go install or Releases above.

### Verify

\`\`\`bash
fly --version
\`\`\`

## Shell completion

Generate completion scripts for your shell:

\`\`\`bash
# Bash
fly completion bash > /etc/bash_completion.d/fly

# Zsh
fly completion zsh > \${fpath[1]}/_fly

# Fish
fly completion fish > ~/.config/fish/completions/fly.fish

# Reload your shell or open a new terminal
\`\`\`

## Environment variables

| Variable | Description |
|----------|-------------|
| \`FFLY_API_URL\` | API base URL (default: production). Set to \`http://localhost:8080\` for local dev. |
| \`FFLY_DEV_EMAIL\` | Email for dev login (with \`fly login --dev\`) |
| \`FFLY_DEV_PASSWORD\` | Password for dev login |
| \`FFLY_CONFIG\` | Path to config file (overrides default location) |

Credentials are stored in \`~/.functionfly/credentials.json\` after \`fly login\`.

## Quick reference

| Command | Description |
|---------|-------------|
| \`fly login\` | Authenticate with FunctionFly |
| \`fly init <name>\` | Create a new function project |
| \`fly dev\` | Run function locally |
| \`fly publish\` | Publish to the registry |
| \`fly logs\` | Stream execution logs |
| \`fly rollback\` | Roll back to a previous version |
| \`fly env\` | Manage environment variables |
| \`fly secrets\` | Manage secrets |

See [Commands Reference](/docs/cli-commands) for full details.
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "cli-commands",
        title: "Commands Reference",
        description: "All CLI commands and options",
        content: `
# Commands Reference

Complete reference for every \`fly\` command.

## Authentication

### \`fly login\`

Authenticate with FunctionFly (OAuth or dev email/password).

\`\`\`bash
fly login                    # Open browser for OAuth
fly login --provider github   # Use GitHub
fly login --provider google  # Use Google
fly login --no-browser       # Print auth URL instead of opening browser
fly login --dev --email admin@example.com  # Dev mode (requires FFLY_API_URL)
\`\`\`

### \`fly whoami\`

Show the currently logged-in user.

### \`fly logout\`

Clear stored credentials.

---

## Function lifecycle

### \`fly init <name>\`

Scaffold a new function project in the current directory.

\`\`\`bash
fly init my-function
fly init my-api --template http-api   # http-api | cron-job | webhook | hello-world
\`\`\`

Creates \`functionfly.jsonc\`, \`main.py\` (or template), and \`test.http\`.

### \`fly dev\`

Run the function locally for development.

\`\`\`bash
fly dev           # Default port 8787
fly dev --port 3000
\`\`\`

### \`fly publish\`

Publish the function to the FunctionFly registry.

\`\`\`bash
fly publish
fly publish --access public
fly publish --access private
fly publish --build        # Build before publishing
fly publish --dry-run      # Validate and bundle without publishing
fly publish --force        # Skip confirmation
fly publish --json         # Output JSON
\`\`\`

### \`fly publish batch\`

Publish multiple functions (each subdirectory with a \`functionfly.jsonc\`).

\`\`\`bash
fly publish batch
fly publish batch --dry-run
fly publish batch --pattern "apps/*/functionfly.jsonc"
\`\`\`

### \`fly update <bump>\`

Bump the version in \`functionfly.jsonc\`.

\`\`\`bash
fly update patch   # 1.0.0 → 1.0.1
fly update minor   # 1.0.0 → 1.1.0
fly update major   # 1.0.0 → 2.0.0
fly update 2.0.0   # Set exact version
\`\`\`

### \`fly test\`

Test your deployed function (invoke and validate response).

\`\`\`bash
fly test
fly test --json
\`\`\`

### \`fly rollback\`

Roll back to a previous version.

\`\`\`bash
fly rollback              # Previous version
fly rollback --version 1.0.5
fly rollback --force      # Skip confirmation
fly rollback --json
\`\`\`

---

## Logs and stats

### \`fly logs\`

Stream execution logs.

\`\`\`bash
fly logs
fly logs --follow          # Stream in real time
fly logs --tail 100        # Last N lines
fly logs --since 1h        # Logs from last 1 hour
fly logs --level error     # Filter by level (info, warn, error)
fly logs --json
\`\`\`

### \`fly stats\`

View usage statistics (invocations, latency, errors).

\`\`\`bash
fly stats
fly stats --json
\`\`\`

---

## Environment and secrets

### \`fly env\`

Manage environment variables for the published function.

\`\`\`bash
fly env list               # List (values masked)
fly env set KEY=value      # Set one or more
fly env get KEY            # Get value
fly env unset KEY           # Remove (alias: delete, rm)
fly env list --json
\`\`\`

### \`fly secrets\`

Manage secrets (encrypted, not shown in logs or UI).

\`\`\`bash
fly secrets list           # List secret names
fly secrets set API_KEY=sk-xxx
fly secrets unset API_KEY  # Alias: delete, rm
fly secrets list --json
\`\`\`

---

## Scheduling

### \`fly schedule\`

Manage scheduled (cron) executions.

\`\`\`bash
fly schedule set "*/5 * * * *"     # Every 5 minutes
fly schedule set --preset every-hour
fly schedule list
fly schedule get
fly schedule remove
fly schedule presets        # List available presets
fly schedule trigger         # Trigger run now
\`\`\`

---

## Global options

Available on every command:

| Option | Description |
|--------|-------------|
| \`--help\`, \`-h\` | Show help |
| \`--version\`, \`-v\` | Show version |
| \`--json\` | Output as JSON (where supported) |

## Exit codes

- \`0\`: Success
- \`1\`: General error (auth, config, API failure)
- \`2\`: Invalid usage (missing args, invalid flags)
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "cli-config",
        title: "Configuration",
        description: "Manifest and config files",
        content: `
# CLI Configuration

## Manifest: \`functionfly.jsonc\`

Every function project has a manifest file. The CLI looks for \`functionfly.jsonc\` or \`functionfly.json\` in the current directory.

### Minimal example

\`\`\`json
{
  "$schema": "https://functionfly.com/schemas/functionfly.json",
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "python3.11",
  "public": true
}
\`\`\`

### Full reference

| Field | Type | Description |
|-------|------|-------------|
| \`name\` | string | **Required.** Function name (slug). |
| \`version\` | string | **Required.** Semver (e.g. \`1.0.0\`). |
| \`runtime\` | string | **Required.** e.g. \`python3.11\`, \`node20\`. |
| \`public\` | boolean | Whether the function is public in the registry (default: \`true\`). |
| \`description\` | string | Short description. |
| \`deterministic\` | boolean | Ensures reproducible builds (default: \`true\`). |
| \`cache_ttl\` | number | Cache TTL in seconds (default: \`86400\`). |
| \`timeout_ms\` | number | Request timeout in milliseconds (default: \`5000\`). |
| \`memory_mb\` | number | Memory limit in MB (default: \`128\`). |
| \`schedule\` | string | Cron expression for scheduled runs (optional). |
| \`env\` | object | Static env key-value pairs (optional). |
| \`dependencies\` | object | Runtime dependencies (optional). |

### Example with options

\`\`\`json
{
  "$schema": "https://functionfly.com/schemas/functionfly.json",
  "name": "my-api",
  "version": "1.0.0",
  "runtime": "python3.11",
  "public": true,
  "description": "A REST API",
  "timeout_ms": 10000,
  "memory_mb": 256,
  "cache_ttl": 3600,
  "schedule": "0 * * * *",
  "env": {
    "LOG_LEVEL": "info"
  }
}
\`\`\`

### JSONC comments

\`functionfly.jsonc\` supports comments:

\`\`\`jsonc
{
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "python3.11",
  // "public": false  // uncomment for private
}
\`\`\`

## Config file (global)

Global CLI config is stored in:

- \`~/.functionfly/config.yaml\` (or path in \`FFLY_CONFIG\`)

Used for API URL and app context when not in a function directory. Credentials are stored separately in \`~/.functionfly/credentials.json\` after \`fly login\`.

## Working directory

All project-scoped commands (\`fly publish\`, \`fly logs\`, \`fly env\`, etc.) must be run from the directory that contains \`functionfly.jsonc\`, or they will fail with a message to run \`fly init\`.
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "agent",
    title: "Agent",
    icon: Cpu,
    pages: [
      {
        slug: "agent-overview",
        title: "Overview",
        description: "Understanding the FunctionFly Agent",
        content: `
# Agent Overview

The FunctionFly Agent is an AI-powered deployment assistant that helps you build, deploy, and manage functions.

## What Can the Agent Do?

### Code Generation
- Generate functions from natural language descriptions
- Create boilerplate for common patterns
- Optimize existing code for edge deployment

### Deployment Management
- Deploy functions with a single command
- Rollback to previous versions
- Manage environment variables

### Troubleshooting
- Analyze error logs
- Suggest performance improvements
- Debug deployment issues

## Agent Modes

### Simple Mode
Best for quick tasks and beginners:
- Natural language commands
- Guided workflows
- Automatic configuration

### Complex Mode
For advanced users and complex scenarios:
- Multi-step planning
- Custom configurations
- Full control over deployment

## Using the Agent

### Via CLI

\`\`\`bash
# Start an interactive session
fly agent

# Run a specific command
fly agent "deploy my API function to production"
\`\`\`

### Via Dashboard

Access the Agent through the web dashboard for a visual interface with:
- File browser
- Real-time logs
- Deployment previews
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "agent-modes",
        title: "Modes",
        description: "Simple and Complex agent modes",
        content: `
# Agent Modes

The FunctionFly Agent operates in two modes to accommodate different use cases.

## Simple Mode

Simple Mode is designed for quick, straightforward tasks.

### When to Use
- Quick deployments
- Simple debugging
- Status checks
- Basic configuration

### Example
\`\`\`bash
$ fly agent "deploy my function"
✓ Building function... done
✓ Deploying to Cloudflare... done
✓ Deploying to Vercel... done
✓ Deployment complete!
\`\`\`

## Complex Mode

Complex Mode is designed for multi-step tasks that require planning.

### When to Use
- Multi-function deployments
- Infrastructure changes
- Database migrations
- Complex debugging

### Example
\`\`\`bash
$ fly agent --complex "set up a new API with authentication"

The agent will:
1. Create authentication middleware
2. Set up API routes
3. Configure environment variables
4. Deploy to staging
5. Run integration tests

Proceed? [Y/n]
\`\`\`

## Switching Modes

### CLI Flag
\`\`\`bash
fly agent --complex    # Use complex mode
fly agent --simple     # Use simple mode (default)
\`\`\`

### Configuration
Set in \`functionfly.jsonc\`:
\`\`\`json
{
  "agent": {
    "defaultMode": "complex"
  }
}
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "agent-ai-llm-integration",
        title: "AI & LLM integration",
        description: "Let LLMs and AI agents discover and call your public functions",
        content: `
# AI & LLM integration

FunctionFly exposes the public registry to AI agents via a standard discovery endpoint. LLMs (OpenAI, Anthropic, Gemini, open-source agents) can fetch one URL, discover all public functions, and call them using native tool-calling APIs—no custom integration code.

## Discovery endpoint

\`\`\`
GET /.well-known/functionfly.json
\`\`\`

- **Public** — no authentication required
- **Cacheable** — \`Cache-Control: public, max-age=300\` (5 minutes)
- **Content-Type** — \`application/json\`

### Query parameters (all optional)

| Param     | Type   | Description                    |
|-----------|--------|--------------------------------|
| \`category\` | string | Filter by function category    |
| \`tags\`     | string | Comma-separated tag filter     |
| \`author\`   | string | Filter by author               |
| \`q\`        | string | Text search                    |
| \`limit\`    | int    | Max functions (default 50, max 200) |
| \`offset\`   | int    | Pagination offset              |

## Response

The response includes \`schema_version\`, \`provider\`, \`api_base\`, \`execution_endpoint\`, \`agent_endpoint\`, \`discovery_endpoint\`, \`generated_at\`, \`total_functions\`, and a \`functions\` array. Each function has:

- \`uri\`, \`name\`, \`title\`, \`description\`, \`version\`, \`category\`, \`tags\`
- \`execution_url\`, \`agent_execution_url\` — where to call the function
- \`tool_schema\` — **OpenAI-compatible** tool definition (name, description, parameters) for use in \`tools\` / \`tool_choice\`

## Example: OpenAI

\`\`\`python
import openai
import requests

# Discover all FunctionFly functions
manifest = requests.get("https://api.functionfly.com/.well-known/functionfly.json").json()

# Extract tool schemas for OpenAI
tools = [fn["tool_schema"] for fn in manifest["functions"]]

# Use in chat completion
response = openai.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Slugify this: Hello World"}],
    tools=tools,
    tool_choice="auto"
)
\`\`\`

## Example: Anthropic

\`\`\`typescript
const manifest = await fetch("https://api.functionfly.com/.well-known/functionfly.json").then(r => r.json());
const tools = manifest.functions.map(fn => fn.tool_schema.function);

const response = await anthropic.messages.create({
  model: "claude-opus-4-5",
  tools: tools,
  messages: [{ role: "user", content: "Convert 'hello world' to a slug" }]
});
\`\`\`

## Calling a function

After the LLM chooses a tool (function), call it via:

- **Direct execution:** \`POST /v1/fx/{author}/{name}\` with \`{"input": <args>}\`
- **Agent execution (quota/attribution):** \`POST /v1/agent/execute/{author}/{name}\` with \`X-Agent-API-Key\` and \`{"input": <args>}\`

The \`tool_schema\` \`parameters\` match the function's input schema so the model can fill arguments correctly.
        `,
        lastUpdated: "2026-03-01"
      }
    ]
  },
  {
    id: "deployment",
    title: "Deployment",
    icon: Cloud,
    pages: [
      {
        slug: "deployment-overview",
        title: "Overview",
        description: "Understanding multi-provider deployment",
        content: `
# Deployment Overview

FunctionFly's deployment system is designed to make multi-provider deployment as simple as single-provider deployment.

## How It Works

1. **Build**: Your code is built and bundled
2. **Transform**: Provider-specific adapters are applied
3. **Deploy**: Deployed to all configured providers simultaneously
4. **Verify**: Health checks ensure successful deployment
5. **Route**: Traffic is routed to the new version

## Deployment Flow

\`\`\`
Your Code → Build → Transform → Deploy to All Providers → Health Check → Live
\`\`\`

## Configuration

Configure providers in \`functionfly.jsonc\`:

\`\`\`json
{
  "providers": {
    "vercel": {
      "enabled": true,
      "regions": ["iad1", "sfo1"]
    },
    "cloudflare": {
      "enabled": true
    },
    "fly": {
      "enabled": true,
      "regions": ["iad", "lax"]
    },
    "deno": {
      "enabled": true
    }
  }
}
\`\`\`

## Deployment Strategies

### All Providers (Default)
Deploy to all enabled providers simultaneously.

### Rolling
Deploy to one provider at a time, verifying each.

### Canary
Deploy to a subset of traffic on each provider.

## Environment Variables

Set environment variables for each provider:

\`\`\`json
{
  "env": {
    "API_KEY": "\${API_KEY}",
    "DATABASE_URL": "\${DATABASE_URL}"
  }
}
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "cli",
        title: "CLI Reference",
        description: "Deployment and CLI quick reference",
        content: `
# CLI Reference (Deployment)

For the complete CLI documentation, see the dedicated **CLI** section:

- [CLI Overview](/docs/cli-overview) – Installation, shell completion, environment variables
- [Commands Reference](/docs/cli-commands) – Every \`fly\` command and option
- [Configuration](/docs/cli-config) – \`functionfly.jsonc\` manifest and config files

## Quick deployment commands

\`\`\`bash
fly login              # Authenticate
fly init <name>        # Create a function
fly dev                # Run locally
fly publish            # Publish to registry
fly logs --follow      # Stream logs
fly rollback           # Roll back
fly env set KEY=value  # Set env vars
\`\`\`

## Manifest

Deployment is driven by \`functionfly.jsonc\` in your project directory. See [CLI Configuration](/docs/cli-config) for the full manifest reference.
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "env-vars",
        title: "Environment Variables",
        description: "Managing secrets and environment configuration",
        content: `
# Environment Variables

Manage configuration and secrets across all providers from a single place.

## Setting Variables

### Via CLI

\`\`\`bash
# Set a variable (synced to all providers)
fly env set API_KEY=sk_live_xxx

# Set for a specific provider
fly env set DATABASE_URL=postgres://... --provider=vercel

# List all variables
fly env list

# Remove a variable
fly env rm API_KEY
\`\`\`

### Via Dashboard

1. Open your function in the dashboard
2. Go to **Settings** → **Environment variables**
3. Add or edit variables; they are encrypted at rest and synced to all providers on next deploy

## Scopes

| Scope | When applied |
|-------|--------------|
| **All providers** | Default; variable is available on every deployment |
| **Per provider** | Use \`--provider=vercel\` (or cloudflare, fly, deno) |
| **Per environment** | Use \`--env=production\` or \`--env=staging\` |

## Security

- **Secrets**: Values are encrypted at rest (AES-256) and in transit (TLS 1.3)
- **Masking**: Secret values are never shown in logs or the UI after creation
- **Rotation**: Rotate keys in the dashboard; redeploy to push new values to providers

## Reference in Code

\`\`\`python
import os

# All env vars are available as standard environment variables
api_key = os.environ.get("API_KEY")
db_url = os.environ.get("DATABASE_URL")
\`\`\`

## Best Practices

1. Use the dashboard or CLI for secrets; avoid committing \`.env\` files
2. Use different values per environment (staging vs production)
3. Document required variables in your function's README
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "cicd",
        title: "CI/CD",
        description: "Deploy from GitHub Actions, GitLab CI, and more",
        content: `
# CI/CD Integration

Deploy FunctionFly functions from your existing CI/CD pipelines.

## GitHub Actions

\`\`\`yaml
name: Deploy to FunctionFly

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"

      - name: Install fly CLI
        run: go install github.com/functionfly/functionfly/cmd/fly@latest

      - name: Add Go bin to PATH
        run: echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      - name: Deploy
        env:
          FUNCTIONFLY_API_KEY: \${{ secrets.FUNCTIONFLY_API_KEY }}
        run: fly deploy
\`\`\`

## GitLab CI

\`\`\`yaml
deploy:
  stage: deploy
  image: golang:1.24
  script:
    - go install github.com/functionfly/functionfly/cmd/fly@latest
    - export PATH="$(go env GOPATH)/bin:$PATH"
    - fly deploy
  variables:
    FUNCTIONFLY_API_KEY: \$FUNCTIONFLY_API_KEY
  only:
    - main
\`\`\`

## API Key for CI

1. In the dashboard: **Settings** → **API keys** → **Create key**
2. Name it (e.g. \`ci-github\`) and copy the key
3. Store as a secret in your CI system: \`FUNCTIONFLY_API_KEY\`
4. Configure non-interactive credentials for CI (e.g. API key env vars) when \`fly login\` is not possible

## Options

| Flag | Description |
|------|-------------|
| \`--json\` | Machine-readable deploy output (useful in CI) |
| \`--skip-tests\` | Skip pre-deploy tests |
| \`--provider\` | Deploy only to one provider |
| \`--env\` | Target environment (e.g. staging) |

## Rollbacks

To rollback from CI, use the same API key and run:

\`\`\`bash
fly rollback --version <previous-version>
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "rollbacks",
        title: "Rollbacks",
        description: "Revert to a previous deployment",
        content: `
# Rollbacks

Revert to a previous deployment across all providers when needed.

## When to Rollback

- A bad release causes errors or downtime
- A provider-specific issue appears after deploy
- You need to quickly restore a known-good state

## Via CLI

\`\`\`bash
# List recent deployments (versions)
fly releases list

# Rollback to the previous deployment
fly rollback

# Rollback to a specific version
fly rollback --version 20260226120000

# Rollback only one provider
fly rollback --provider=vercel
\`\`\`

## Via Dashboard

1. Open your function
2. Go to **Deployments** or **Releases**
3. Find the deployment you want to restore
4. Click **Rollback** and confirm

## What Happens

- Traffic is switched to the selected previous version
- All providers are updated to that version (or only the selected provider)
- No new build is run; the previous artifact is re-used
- Rollback typically completes in under a minute

## After a Rollback

- Investigate the failed release (logs, metrics) before deploying again
- Use [Troubleshooting](/docs/common-errors) for common causes
- Consider enabling [canary deployments](/docs/deployment-overview#deployment-strategies) for riskier changes
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "state-fabric",
    title: "State Fabric",
    icon: Database,
    pages: [
      {
        slug: "state-fabric-overview",
        title: "Overview",
        description: "Stateful capabilities for serverless",
        content: `
# State Fabric Overview

State Fabric brings stateful capabilities to your serverless functions without managing infrastructure.

## Features

### KV Store
Simple key-value storage with edge caching:

\`\`\`python
from functionfly import kv

# Set a value
await kv.set("user:123", {"name": "John", "role": "admin"})

# Get a value
user = await kv.get("user:123")

# Delete
await kv.delete("user:123")
\`\`\`

### Pub/Sub
Real-time messaging between functions:

\`\`\`python
from functionfly import pubsub

# Publish
await pubsub.publish("orders", {"id": "123", "total": 99.99})

# Subscribe
async for message in pubsub.subscribe("orders"):
    print(f"New order: {message}")
\`\`\`

### Queues
Reliable message queuing:

\`\`\`python
from functionfly import queue

# Enqueue
await queue.enqueue("emails", {"to": "user@example.com", "template": "welcome"})

# Dequeue
job = await queue.dequeue("emails")
\`\`\`

### Sessions
User session management:

\`\`\`python
from functionfly import sessions

# Create session
session = await sessions.create({"userId": "123"})

# Get session
data = await sessions.get(session.id)

# Update
await sessions.update(session.id, {"lastSeen": "2024-01-01"})
\`\`\`

## Consistency Models

Choose the right consistency model for your use case:

| Model | Latency | Consistency | Use Case |
|-------|---------|-------------|----------|
| Eventual | Low | Eventual | Caching, counters |
| Strong | Medium | Strong | User data, transactions |
| Session | Low | Session-scoped | Shopping carts, forms |
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "security",
    title: "Security",
    icon: Shield,
    pages: [
      {
        slug: "security-overview",
        title: "Overview",
        description: "Security features and best practices",
        content: `
# Security Overview

FunctionFly is built with security as a core principle.

## Security Features

### Authentication
- OAuth 2.0 / OpenID Connect
- JWT tokens with automatic rotation
- Multi-factor authentication (MFA)
- API key management

### Authorization
- Role-based access control (RBAC)
- Fine-grained permissions
- Resource-level policies
- Audit logging

### Encryption
- TLS 1.3 for all traffic
- AES-256 encryption at rest
- Environment variable encryption
- Secure key management

### Network Security
- DDoS protection via Cloudflare
- IP allowlisting
- VPC isolation for databases
- Private networking options

## Compliance

FunctionFly maintains compliance with:

- SOC 2 Type II
- GDPR
- HIPAA (Business Associate Agreement available)
- PCI DSS (for payment processing)

## Best Practices

### Function Security
1. Validate all inputs
2. Use parameterized queries
3. Sanitize user-generated content
4. Implement rate limiting
5. Log security events

### API Security
\`\`\`python
from functionfly import security

@security.rate_limit(max_requests=100, window=60)
def handle_request(req):
    # Validate input
    security.validate_json_schema(req.body, schema)

    # Check authentication
    user = security.require_auth(req)

    # Check authorization
    security.require_permission(user, "functions:write")

    return {"success": True}
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "authentication",
        title: "Authentication",
        description: "OAuth, JWT, and MFA",
        content: `
# Authentication

FunctionFly supports multiple authentication methods for the dashboard and API.

## Dashboard Login

- **Email + password** with optional MFA
- **OAuth**: Sign in with Google, GitHub, or other configured IdPs
- **SSO**: SAML 2.0 and OpenID Connect for enterprise

## API Authentication

### API Keys

Best for server-to-server and CI/CD:

\`\`\`bash
curl -H "Authorization: Bearer ff_sk_xxx" \\
  https://api.functionfly.io/v1/functions
\`\`\`

Create and manage keys in **Settings** → **API keys**. Keys can be scoped to a team or function.

### OAuth 2.0

For third-party apps and user-delegated access:

- **Authorization code** with PKCE for public clients
- **Client credentials** for machine-to-machine
- Tokens are JWTs with short-lived access and optional refresh

### JWT

- Access tokens expire in 1 hour (configurable)
- Refresh tokens (if enabled) last 30 days
- Token rotation on refresh for security

## MFA (Multi-Factor Authentication)

1. Enable in **Settings** → **Security**
2. Use an authenticator app (TOTP) or hardware key
3. Required for sensitive actions (e.g. delete function, change billing) when enabled

## Best Practices

- Use API keys for automation; rotate them periodically
- Enable MFA on accounts with deploy or admin access
- Prefer short-lived tokens and refresh when building integrations
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "api-keys",
        title: "API Keys",
        description: "Create and manage API keys",
        content: `
# API Keys

API keys authenticate requests to the FunctionFly API and CLI (e.g. in CI).

## Creating a Key

### Dashboard

1. Go to **Settings** → **API keys**
2. Click **Create key**
3. Name the key (e.g. \`Production CI\`, \`Staging\`)
4. Optionally restrict scope (all functions vs specific function or team)
5. Copy the key once; it is not shown again

### CLI

\`\`\`bash
fly api-keys create --name "CI key" --scope function:my-app
\`\`\`

## Using a Key

### Environment variable

\`\`\`bash
export FUNCTIONFLY_API_KEY=ff_sk_xxx
fly deploy
\`\`\`

### Header

\`\`\`
Authorization: Bearer ff_sk_xxx
\`\`\`

## Scopes

| Scope | Access |
|-------|--------|
| Full | All functions and resources in the account |
| Team | Only functions in the selected team(s) |
| Function | Single function (read/deploy/logs) |

## Rotation

1. Create a new key with the same scope
2. Update CI/secrets to use the new key
3. Verify deployments work
4. Revoke the old key in **Settings** → **API keys**

## Security

- Keys are stored hashed; the plain value is shown only at creation
- Revoke keys immediately if they are exposed
- Prefer narrow scopes (e.g. one function) for CI keys
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "monitoring",
    title: "Monitoring",
    icon: Gauge,
    pages: [
      {
        slug: "monitoring-overview",
        title: "Overview",
        description: "Observability and analytics",
        content: `
# Monitoring Overview

FunctionFly provides comprehensive observability into your functions.

## Metrics

### Function Metrics
- **Invocations**: Total number of calls
- **Duration**: Execution time (p50, p95, p99)
- **Errors**: Error rates and types
- **Cold Starts**: Cold start frequency and duration

### Provider Metrics
- **Availability**: Uptime by provider
- **Latency**: Response times by region
- **Throughput**: Requests per second

## Logging

### Structured Logging
\`\`\`python
from functionfly import log

log.info("Processing payment", {
    "order_id": "123",
    "amount": 99.99,
    "user_id": "user_456"
})

log.error("Payment failed", {
    "order_id": "123",
    "error": "insufficient_funds"
})
\`\`\`

### Log Levels
- DEBUG: Detailed debugging information
- INFO: General operational information
- WARN: Warning events
- ERROR: Error events
- CRITICAL: Critical failures

## Alerting

Configure alerts based on metrics:

\`\`\`yaml
alerts:
  - name: high_error_rate
    condition: error_rate > 5%
    duration: 5m
    severity: critical

  - name: slow_response
    condition: p95_latency > 500ms
    duration: 10m
    severity: warning
\`\`\`

## Tracing

Distributed tracing across providers:

\`\`\`python
from functionfly import trace

@trace.span("process_order")
async def process_order(order_id):
    with trace.span("validate_order"):
        await validate(order_id)

    with trace.span("charge_payment"):
        await charge(order_id)

    with trace.span("send_confirmation"):
        await notify(order_id)
\`\`\`

## Dashboards

Access pre-built dashboards for:
- Overview: High-level health metrics
- Performance: Latency and throughput
- Errors: Error analysis and trends
- Costs: Spend by provider and function
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "alerts",
        title: "Alerts",
        description: "Configure alerts and notifications",
        content: `
# Alerts

Configure alerts so you're notified when metrics cross thresholds.

## Alert Types

### Function-level

- **Error rate** above a percentage (e.g. > 5%)
- **Latency** above a threshold (e.g. p95 > 500ms)
- **Invocations** spike or drop
- **Cold starts** frequency

### Account-level

- **Quota** (invocations, bandwidth) approaching limit
- **Billing** threshold

## Creating an Alert

### Dashboard

1. Open **Monitoring** → **Alerts**
2. Click **Create alert**
3. Choose metric, condition (e.g. \`error_rate > 5%\`), and duration (e.g. 5 minutes)
4. Add channels: email, Slack, PagerDuty, or webhook

### Configuration file

\`\`\`yaml
# functionfly.alerts.yaml (optional)
alerts:
  - name: high_error_rate
    metric: error_rate
    condition: "> 0.05"
    window: 5m
    severity: critical
    channels:
      - type: slack
        url: \$SLACK_WEBHOOK
  - name: slow_response
    metric: p95_latency_ms
    condition: "> 500"
    window: 10m
    severity: warning
\`\`\`

## Severity

| Level | Use case |
|-------|----------|
| Critical | Page immediately; likely user impact |
| Warning | Investigate soon |
| Info | For awareness (e.g. deploy completed) |

## Webhooks

Send alerts to any HTTP endpoint:

\`\`\`json
{
  "event": "alert.triggered",
  "alert": "high_error_rate",
  "function": "my-api",
  "value": 0.12,
  "threshold": 0.05,
  "timestamp": "2026-02-26T12:00:00Z"
}
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "logs",
        title: "Logs",
        description: "View and query function logs",
        content: `
# Logs

Access structured logs for debugging and auditing.

## Viewing Logs

### Dashboard

1. Open your function
2. Go to **Logs**
3. Filter by time range, level (info, error), or provider
4. Click a log line for full context (request id, trace id)

### CLI

\`\`\`bash
# Tail logs (live)
fly logs --tail

# Last 100 lines
fly logs -n 100

# Filter by level
fly logs --level error

# Filter by provider
fly logs --provider=cloudflare
\`\`\`

## Log Levels

- **DEBUG**: Detailed diagnostics (disable in production)
- **INFO**: Normal operations (e.g. request completed)
- **WARN**: Recoverable issues (e.g. retry)
- **ERROR**: Failures (e.g. exception)
- **CRITICAL**: Severe failure (e.g. crash)

## Structured Logging

\`\`\`python
from functionfly import log

log.info("order_created", order_id="123", amount=99.99, user_id="u_456")
log.error("payment_failed", order_id="123", reason="insufficient_funds")
\`\`\`

Query in the dashboard with \`order_id:123\` or \`level:error\`.

## Retention

- **Free**: 7 days
- **Pro**: 30 days
- **Enterprise**: Configurable (e.g. 90 days or export to your SIEM)

## Best Practices

1. Use structured fields (not only message) for filtering
2. Avoid logging secrets or PII
3. Use appropriate levels so alerts stay actionable
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "api-reference",
    title: "API Reference",
    icon: FileText,
    pages: [
      {
        slug: "api-overview",
        title: "Overview",
        description: "REST API introduction",
        content: `
# API Reference Overview

The FunctionFly REST API lets you manage functions, deployments, and resources programmatically.

## Base URL

\`\`\`
https://api.functionfly.io/v1
\`\`\`

## Authentication

All requests require authentication via API key or OAuth token:

\`\`\`
Authorization: Bearer <token>
\`\`\`

See [Authentication](/docs/authentication) and [API Keys](/docs/api-keys) for details.

## Rate Limits

| Plan | Requests/minute |
|------|-----------------|
| Free | 60 |
| Pro | 300 |
| Enterprise | Custom |

\`429 Too Many Requests\` is returned when exceeded. Retry after the \`Retry-After\` header.

## Pagination

List endpoints return paginated results:

\`\`\`
GET /v1/functions?page=2&per_page=20
\`\`\`

Response includes \`page\`, \`per_page\`, \`total\`, and \`data\`.

## Errors

Errors use a consistent format:

\`\`\`json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid request body",
    "details": { "field": "name" }
  }
}
\`\`\`

Common HTTP codes: \`400\` (bad request), \`401\` (unauthorized), \`404\` (not found), \`429\` (rate limit).
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "rest-api",
        title: "REST API",
        description: "Endpoints for functions and deployments",
        content: `
# REST API Endpoints

## Functions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | \`/v1/functions\` | List functions |
| POST | \`/v1/functions\` | Create function |
| GET | \`/v1/functions/:id\` | Get function |
| PATCH | \`/v1/functions/:id\` | Update function |
| DELETE | \`/v1/functions/:id\` | Delete function |

## Deployments

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | \`/v1/functions/:id/deployments\` | Create deployment |
| GET | \`/v1/functions/:id/deployments\` | List deployments |
| GET | \`/v1/functions/:id/deployments/:vid\` | Get deployment |
| POST | \`/v1/functions/:id/rollback\` | Rollback to version |

## Environment Variables

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | \`/v1/functions/:id/env\` | List env vars (values masked) |
| PUT | \`/v1/functions/:id/env\` | Set env vars (bulk) |

## Invoke

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | \`/v1/functions/:id/invoke\` | Invoke function (sync) |

## Example

\`\`\`bash
# List functions
curl -H "Authorization: Bearer \$TOKEN" \\
  https://api.functionfly.io/v1/functions

# Create deployment
curl -X POST -H "Authorization: Bearer \$TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"ref": "main"}' \\
  https://api.functionfly.io/v1/functions/fn_xxx/deployments
\`\`\`
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  },
  {
    id: "troubleshooting",
    title: "Troubleshooting",
    icon: AlertTriangle,
    pages: [
      {
        slug: "troubleshooting-overview",
        title: "Overview",
        description: "Common issues and how to resolve them",
        content: `
# Troubleshooting Overview

Quick links to the most common issues and where to get help.

## Common Issues

- [Deploy failures](/docs/common-errors#deploy-failures) – Build errors, provider timeouts, config issues
- [Runtime errors](/docs/common-errors#runtime-errors) – Function crashes, timeouts, out of memory
- [Routing and failover](/docs/common-errors#routing) – Traffic not reaching the right provider
- [Authentication](/docs/common-errors#authentication) – Login, API keys, OAuth

## Debugging

- [Enable debug logging](/docs/debugging) – CLI and SDK
- [Inspect logs and traces](/docs/logs) – Dashboard and CLI
- [Reproduce locally](/docs/debugging#local-reproduction) – \`fly dev\` and provider simulators

## Getting Help

- [Contact Support](/contact) – For account and platform issues
- [FAQ](/faq) – Common questions
- [Status page](https://status.functionfly.com) – Outages and incidents
- [Community](https://github.com/functionfly/functionfly/discussions) – Discussions and examples
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "common-errors",
        title: "Common Errors",
        description: "Frequent errors and solutions",
        content: `
# Common Errors

## Deploy Failures

### \`Build failed: module not found\`

- Ensure all dependencies are in \`package.json\` (Node) or \`requirements.txt\` (Python)
- Run \`fly build\` locally to reproduce
- Check [CLI Commands](/docs/cli-commands) and [Configuration](/docs/cli-config) for supported runtimes

### \`Provider timeout: Vercel\` (or Cloudflare, Fly, Deno)

- Large bundles or slow builds can hit provider limits
- Reduce bundle size (tree-shake, split) or increase timeout in \`functionfly.jsonc\`
- Deploy to one provider first: \`fly deploy --provider=cloudflare\`

### \`Invalid configuration\`

- Validate \`functionfly.jsonc\` with \`fly config validate\`
- Ensure \`runtime\`, \`name\`, and \`providers\` are set correctly

## Runtime Errors

### \`Function timeout\`

- Default timeout is 30s (varies by provider)
- Increase in config or optimize long-running work (use queues, background jobs)

### \`Out of memory\`

- Increase memory in \`functionfly.jsonc\` (e.g. \`memory: 512\`)
- Profile with [Monitoring](/docs/monitoring-overview) to find leaks or heavy allocations

### \`Cold start timeout\`

- Optimize startup (lazy load, reduce dependencies)
- Use [warm-up](/docs/deployment-overview) if available for your plan

## Routing

### Traffic not failing over

- Check [provider status](https://status.functionfly.com)
- Verify health checks in the dashboard (Deployments → Health)
- Ensure at least two providers are enabled and deployed

## Authentication

### \`Invalid API key\`

- Create a new key in **Settings** → **API keys** and update your client
- Ensure no extra spaces; use \`Authorization: Bearer <key>\`

### \`Unauthorized\` on dashboard

- Clear cookies or try incognito; re-login
- If SSO: confirm your IdP is configured and your account is linked
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "debugging",
        title: "Debugging",
        description: "Enable debug logs and reproduce issues",
        content: `
# Debugging

## Enable Debug Logging

### CLI

\`\`\`bash
fly deploy --debug
# or
export FUNCTIONFLY_DEBUG=1
fly deploy
\`\`\`

### SDK (in your function)

\`\`\`python
import os
os.environ["FUNCTIONFLY_LOG_LEVEL"] = "DEBUG"
\`\`\`

## Local Reproduction

\`\`\`bash
# Run function locally (simulates request/response)
fly dev

# Invoke with custom body
curl -X POST http://localhost:3000/ -d '{"key": "value"}'
\`\`\`

Match runtime and env vars to production when possible.

## Inspect Logs and Traces

1. **Dashboard** → Your function → **Logs**: filter by time, level, request id
2. **CLI**: \`fly logs --tail\` for live logs
3. Use the **request id** from responses to trace a single request across logs and [tracing](/docs/monitoring-overview#tracing)

## Provider-Specific Debugging

- **Vercel**: Check Vercel dashboard for build/runtime logs
- **Cloudflare**: Workers dashboard → Logs and Metrics
- **Fly.io**: \`fly logs\` in your Fly app
- **Deno**: Deploy dashboard logs

## Reporting a Bug

When contacting support, include:

- Function name and (if applicable) deployment version
- Request id or timestamp of the failure
- Relevant log snippet (redact secrets)
- Steps to reproduce
- Expected vs actual behavior
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "getting-help",
        title: "Getting Help",
        description: "Support channels and resources",
        content: `
# Getting Help

## Support Channels

| Channel | Use for | Response |
|---------|---------|----------|
| [Contact form](/contact) | Account, billing, platform bugs | Within 1–2 business days |
| [FAQ](/faq) | Common how-to and conceptual questions | Self-serve |
| [Status page](https://status.functionfly.com) | Outages and incidents | Live updates |
| [GitHub Discussions](https://github.com/functionfly/functionfly/discussions) | Community help, examples, feature ideas | Community and team |

## Before You Contact Support

1. Check [Common Errors](/docs/common-errors) and [FAQ](/faq)
2. Search [GitHub Discussions](https://github.com/functionfly/functionfly/discussions)
3. Gather: function name, deployment id or timestamp, request id, error message, and steps to reproduce

## Enterprise Support

Enterprise plans include:

- Dedicated support channel (email/Slack)
- SLA-backed response times
- Architecture and best-practice reviews
- See [Enterprise Support](/enterprise/support) for details
        `,
        lastUpdated: "2026-02-27"
      }
    ]
  }
];

// Helper to find a page by slug
export function findPageBySlug(slug: string): DocPage | undefined {
  for (const section of docSections) {
    const page = section.pages.find(p => p.slug === slug);
    if (page) return page;
  }
  return undefined;
}

// Helper to get all pages flattened
export function getAllPages(): DocPage[] {
  return docSections.flatMap(section => section.pages);
}

// Helper to search pages
export function searchPages(query: string): DocPage[] {
  const lowerQuery = query.toLowerCase();
  return getAllPages().filter(page =>
    page.title.toLowerCase().includes(lowerQuery) ||
    page.description.toLowerCase().includes(lowerQuery) ||
    page.content.toLowerCase().includes(lowerQuery)
  );
}
