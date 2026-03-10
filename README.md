# FunctionFly

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/functionfly/orchestrator)](https://hub.docker.com/r/functionfly/orchestrator)
[![Discord](https://img.shields.io/discord/123456789?label=Discord)](https://discord.gg/functionfly)

A production-ready serverless function platform built with Go, designed for high-performance function execution at the edge.

</div>

## Overview

FunctionFly is a comprehensive serverless platform that enables developers to deploy and run functions in various languages with automatic scaling, built-in monitoring, and a pay-per-use pricing model.

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
| `cmd/` | Go entrypoints (orchestrator-api, fly CLI, health-monitor, etc.) |
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
- PostgreSQL 14+
- Redis 7+

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/functionfly/functionfly.git
cd functionfly
```

2. Copy environment configuration:
```bash
cp .env.example .env
```

3. Start the development environment:
```bash
docker-compose -f docker-compose.local.yml up -d
```

4. Run the orchestrator:
```bash
go run cmd/fly/main.go serve
```

5. Deploy your first function:
```bash
go run cmd/fly/main.go deploy --path ./examples/hello-world
```

### Using the CLI

```bash
# Login to FunctionFly
fly login

# List your functions
fly list

# Deploy a function
fly deploy --name my-function --path ./my-function

# Invoke a function
fly invoke my-function --data '{"name": "World"}'

# View logs
fly logs my-function
```

## Deployment

### Docker Compose (Production)

```bash
cp .env.production.example .env.production
# Edit .env.production with your production values

docker-compose -f docker-compose.production.yml up -d
```

### Kubernetes

See [deploy/kubernetes](./deploy/kubernetes) for Kubernetes manifests.

### Cloud Providers

- **AWS**: [deploy/aws](./deploy/aws)
- **GCP**: [deploy/gcp](./deploy/gcp)

## SDKs

- **Python**: [sdk/python](./sdk/python)
- **Node.js**: [sdk/nodejs](./sdk/nodejs)
- **Go**: [sdk/go](./sdk/go)

### Example: Python SDK

```python
from functionfly import FunctionClient

client = FunctionClient(api_key="your-api-key")

# Invoke a function
result = client.invoke("my-function", {"name": "World"})
print(result)
```

### Example: Node.js SDK

```javascript
const { FunctionClient } = require('@functionfly/sdk');

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

FunctionFly includes built-in monitoring with Prometheus and Grafana.

1. Start the monitoring stack:
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

2. Access Grafana at http://localhost:3000 (admin/admin)

3. Import dashboards from `deploy/monitoring/grafana/`

## Examples

See the [examples](./examples) directory for sample functions:

- [hello-world](./examples/hello-world) - Basic function
- [http-api](./examples/http-api) - HTTP API with routing
- [webhook-notifier](./examples/webhook-notifier) - Webhook handler
- [ai-sentiment](./examples/ai-sentiment) - AI/ML inference

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a Pull Request

## Documentation

- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Security Policy](docs/SECURITY.md)
- [Migration Guide](docs/MIGRATIONS.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- Discord: https://discord.gg/functionfly
- Issues: https://github.com/functionfly/functionfly/issues
