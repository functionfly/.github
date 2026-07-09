# FunctionFly™

<div align="center">

![FunctionFly Logo](./functionfly.png)

**A production-ready serverless function platform with ML-powered intelligence for high-performance execution at the edge.**

[![Discord](https://img.shields.io/badge/Discord-Join-5865F2?logo=discord&logoColor=white)](https://discord.com/invite/cSTsz3WjpD)

</div>

---

## What We Build

FunctionFly™ is a serverless platform that enables developers to deploy and run functions in multiple languages with automatic scaling, built-in monitoring, pay-per-use pricing, and an integrated ML intelligence layer (FlyMind).

### Core Capabilities

- **Multi-language Support**: Go, Python, Node.js, Rust, TypeScript, C, Kotlin, Ruby, Swift, and more
- **Multiple Runtimes**: WASM for sandboxed execution, MicroVM for full environments, Local for development
- **Edge Execution**: Run functions close to your users with global distribution
- **Automatic Scaling**: Scale from zero to millions of requests without configuration
- **Built-in Monitoring**: Real-time metrics with Prometheus and Grafana
- **Agent Runtime**: Stateful Agent Runtime (SAR) for complex agent workflows
- **ML Intelligence (FlyMind)**: Cost anomaly detection, demand forecasting, Thompson Sampling routing, collaborative filtering recommendations
- **MCP Server**: Work with functions from any MCP-compatible client

### Use Cases

- API backends and webhook handlers
- AI/ML inference pipelines
- Real-time data processing
- Multi-step agent workflows
- Custom business logic at the edge

## Open Source Projects

| Repository | Description |
|------------|-------------|
| [ff-cli](https://github.com/functionfly/ff-cli) | Official CLI (`ff`) for FunctionFly |
| [mcp-server](https://github.com/functionfly/mcp-server) | MCP server for functions, agents, vaults, and workflows |
| [homebrew-tap](https://github.com/functionfly/homebrew-tap) | Homebrew tap for macOS/Linux installation |

## Get Started

```bash
# Install the CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.sh | bash

# Deploy your first function
ff login
ff deploy --path ./my-function
```

## Contact

- **Website**: [functionfly.com](https://functionfly.com)
- **Discord**: [discord.com/invite/cSTsz3WjpD](https://discord.com/invite/cSTsz3WjpD)
- **Email**: support@functionfly.com
- **Docs**: [docs.functionfly.com](https://docs.functionfly.com)

---

*© 2026 FunctionFly Inc. — All rights reserved.*