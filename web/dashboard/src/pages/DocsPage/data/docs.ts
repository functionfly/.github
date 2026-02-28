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
# Install the CLI
npm install -g @functionfly/cli

# Login to your account
flypy login

# Create a new function
flypy init my-function

# Deploy to all providers
flypy deploy
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
- [CLI Reference](/docs/cli) - Learn the command-line interface
        `,
        lastUpdated: "2026-02-27"
      },
      {
        slug: "quickstart",
        title: "Quickstart",
        description: "Deploy your first function in under 5 minutes",
        content: `
# Quickstart Guide

Get your first function deployed to multiple edge providers in minutes.

## Prerequisites

- Node.js 18+ installed
- A FunctionFly account (sign up at [functionfly.io](https://functionfly.io))

## Step 1: Install the CLI

\`\`\`bash
npm install -g @functionfly/cli
\`\`\`

## Step 2: Authenticate

\`\`\`bash
flypy login
\`\`\`

This will open a browser window for authentication.

## Step 3: Create a Function

\`\`\`bash
flypy init hello-world
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
flypy deploy
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
flypy agent

# Run a specific command
flypy agent "deploy my API function to production"
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
$ flypy agent "deploy my function"
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
$ flypy agent --complex "set up a new API with authentication"

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
flypy agent --complex    # Use complex mode
flypy agent --simple     # Use simple mode (default)
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
        description: "Command-line interface documentation",
        content: `
# CLI Reference

The FunctionFly CLI (\`flypy\`) is your primary tool for managing functions.

## Installation

\`\`\`bash
npm install -g @functionfly/cli
\`\`\`

## Commands

### Authentication

\`\`\`bash
flypy login          # Log in to FunctionFly
flypy logout         # Log out
flypy whoami         # Show current user
\`\`\`

### Function Management

\`\`\`bash
flypy init <name>           # Create a new function
flypy dev                   # Run local development server
flypy deploy                # Deploy function
flypy logs                  # View function logs
flypy status                # Check deployment status
\`\`\`

### Environment Variables

\`\`\`bash
flypy env list              # List environment variables
flypy env set KEY=value     # Set environment variable
flypy env get KEY           # Get environment variable
flypy env rm KEY            # Remove environment variable
\`\`\`

### Provider Management

\`\`\`bash
flypy providers list        # List configured providers
flypy providers add <name>  # Add a provider
flypy providers rm <name>   # Remove a provider
\`\`\`

## Global Options

\`\`\`bash
--help, -h        Show help
--version, -v     Show version
--debug           Enable debug logging
--config <path>   Use specific config file
\`\`\`

## Configuration File

Default configuration is read from \`functionfly.jsonc\`:

\`\`\`json
{
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "python3.11",
  "providers": {
    "vercel": { "enabled": true },
    "cloudflare": { "enabled": true }
  }
}
\`\`\`
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
