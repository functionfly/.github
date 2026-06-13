# FunctionFly™

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/functionfly/orchestrator)](https://hub.docker.com/r/functionfly/orchestrator)
[![Discord](https://img.shields.io/discord/123456789?label=Discord)](https://discord.gg/functionfly)

A production-ready serverless function platform built with Go, designed for high-performance function execution at the edge.

</div>

## Overview

FunctionFly™ is a comprehensive serverless platform that enables developers to deploy and run functions in various languages with automatic scaling, built-in monitoring, and a pay-per-use pricing model.

## Features

- **Multi-language Support**: Deploy functions in Go, Python, Node.js, and more
- **Edge Execution**: Run functions close to your users with global distribution
- **Automatic Scaling**: Scale from zero to millions of requests without configuration
- **Built-in Monitoring**: Real-time metrics with Prometheus and Grafana dashboards
- **Pay-per-use Pricing**: Pay only for what you use with granular billing
- **Secure by Default**: Isolated execution environments with built-in secrets management
- **CLI & SDK**: Deploy functions from the command line or programmatically via SDKs

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Users     │────▶│   Gateway   │────▶│Orchestrator │
└─────────────┘     └─────────────┘     └─────────────┘
                                                │
                    ┌─────────────┐             │
                    │  Database   │◀────────────┤
                    │  (Postgres) │             │
                    └─────────────┘             ▼
                                      ┌─────────────┐
                                      │   Agent     │
                                      │ (Executors) │
                                      └─────────────┘
                                            │
                                      ┌─────────────┐
                                      │  Runtimes   │
                                      │ (WASM/VM)   │
                                      └─────────────┘
```

## Project layout

| Path | Purpose |
|------|---------|
| `cmd/` | Go entrypoints (orchestrator-api, health-monitor, etc.) |
| `internal/` | Go application code |
| `docs/` | Documentation and guides (see [docs/README.md](docs/README.md)) |
| `scripts/` | Dev and ops scripts (dev.sh, migrations, publish, test helpers) |
| `examples/` | Sample functions and fixtures; `examples/stdlib-publish/` = publish payloads for stdlib |
| `examples/fixtures/` | Sample manifests, test inputs, and tiny scripts |
| `assets/` | Images and static assets |
| `migrations/` | Database migrations |
| `web/` | Dashboard and frontends |
| `deploy/` | Deployment configs (Caddy, etc.) |
| `plans/` | Design and planning docs |

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 17
- Redis 7+

### Local Development

1. Clone the repository:

```bash
git clone https://github.com/functionfly/functionfly.git
cd functionfly
```

1. Copy environment configuration:

```bash
cp .env.example .env
```

1. Start the development environment:

```bash
docker-compose -f docker-compose.local.yml up -d
```

1. Run the orchestrator:

```bash
go build -o bin/orchestrator-api ./cmd/orchestrator-api
./bin/orchestrator-api --skip-migrations
```

Or use the Makefile:

```bash
make dev
```

1. Deploy your first function using the [ff CLI](https://github.com/functionfly/ff-cli):

```bash
# Install the CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.sh | bash

# Login and deploy
fly login
fly deploy --path ./examples/hello-world
```

### Using the CLI

The `ff` CLI is now maintained in its own repository: [functionfly/ff-cli](https://github.com/functionfly/ff-cli)

```bash
# Login to FunctionFly™
ff login

# Initialize a new function project
ff init my-function

# Run local development environment
ff dev

# Publish a function to the registry
ff publish

# Deploy a function
ff deploy

# View logs
ff logs my-function
```

Configuration precedence: **environment variables (FF_*)** override **global config** (`~/.ff/config.yaml`). Use `ff config` to view or `ff config reset` to restore defaults. See the [ff-cli repository](https://github.com/functionfly/ff-cli) for full CLI docs.

## Deployment

### Docker Compose (Production)

```bash
cp .env.production.example .env.production
# Edit .env.production with your production values

docker-compose -f docker-compose.production.yml up -d
```

### Kubernetes

Kubernetes deployment documentation is coming soon. For now, see the [Production Deployment Guide](./docs/PRODUCTION_DEPLOYMENT.md) for bare-metal deployment.

### Cloud Providers

- **AWS**: Contact the team for AWS deployment support
- **GCP**: Contact the team for GCP deployment support

## SDKs

- **Python**: [sdk/python](./sdk/python)
- **JavaScript/TypeScript**: [sdk/js](./sdk/js) — contains the `flypy` package
- **Edge**: [sdk/edge](./sdk/edge)

### Example: Python SDK

```python
from functionfly import FunctionClient

client = FunctionClient(api_key="your-api-key")

# Invoke a function
result = client.invoke("my-function", {"name": "World"})
print(result)
```

### Example: JavaScript SDK (flypy)

```javascript
const { FunctionClient } = require('@functionfly/flypy');

const client = new FunctionClient({ apiKey: 'your-api-key' });

const result = await client.invoke('my-function', { name: 'World' });
console.log(result);
```

## Configuration

Environment variables can be configured via `.env` files. See `.env.example` for available options.

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `REDIS_URL` | Redis connection string | redis://localhost:6379 |
| `PORT` | HTTP server port | 8080 |
| `LOG_LEVEL` | Logging level | info |

## Monitoring

FunctionFly™ includes built-in monitoring with Prometheus and Grafana.

1. Start the monitoring stack:

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

1. Access Grafana at <http://localhost:3000> (admin/admin)

2. Import dashboards from `deploy/monitoring/grafana/`

## Examples

See the [examples](./examples) directory for sample functions:

- [ai-sentiment](./examples/ai-sentiment) - AI/ML sentiment analysis
- [email-notification](./examples/email-notification) - Email webhook handler
- [external-api](./examples/external-api) - External API integration
- [file-storage](./examples/file-storage) - File operations
- [kv-counter](./examples/kv-counter) - Key-value state
- [python](./examples/python) - Python runtime example
- [webhook-notifier](./examples/webhook-notifier) - Webhook handler

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a Pull Request

## Documentation

- [Quick Start Guide](./docs/QUICK_START.md) — Get running locally or in production
- [Production Deployment](./docs/PRODUCTION_DEPLOYMENT.md) — Full production deployment
- [Security Policy](./docs/SECURITY.md)
- [Migration Guide](./docs/MIGRATIONS.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- Discord: <https://discord.gg/functionfly>
- Issues: <https://github.com/functionfly/functionfly/issues>
