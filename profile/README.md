# FunctionFly™

<div align="center">

<h1>
<img src="./functionfly.png" alt="FunctionFly" width="48" style="vertical-align: middle;">
FunctionFly™
</h1>

**A production-ready serverless function platform with ML-powered intelligence for high-performance execution at the edge.**

[![Discord](https://img.shields.io/badge/Discord-Join-5865F2?logo=discord&logoColor=white)](https://discord.com/invite/cSTsz3WjpD)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Website](https://img.shields.io/website?url=https://functionfly.com)](https://functionfly.com)

[![ff-cli](https://img.shields.io/github/actions/workflow/status/functionfly/ff-cli/ci.yml?branch=main&label=ff-cli)](https://github.com/functionfly/ff-cli/actions)
[![mcp-server](https://img.shields.io/github/actions/workflow/status/functionfly/mcp-server/ci.yml?branch=main&label=mcp-server)](https://github.com/functionfly/mcp-server/actions)
[![functionfly](https://img.shields.io/github/actions/workflow/status/functionfly/functionfly/ci.yml?branch=develop&label=functionfly)](https://github.com/functionfly/functionfly/actions)

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

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3776AB?style=flat&logo=python&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat&logo=typescript&logoColor=white)
![Rust](https://img.shields.io/badge/Rust-000000?style=flat&logo=rust&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-339933?style=flat&logo=nodedotjs&logoColor=white)
![React](https://img.shields.io/badge/React-61DAFB?style=flat&logo=react&logoColor=black)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)
![WebAssembly](https://img.shields.io/badge/WebAssembly-654FF0?style=flat&logo=webassembly&logoColor=white)

## Community

<img src="https://discordapp.com/api/guilds/1524584614453055548/widget.png?style=300" alt="FunctionFly Discord Server" style="border-radius: 8px; max-width: 100%;">

Join our Discord community to get help, share feedback, and stay up to date.

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

## Roadmap

We are actively building FunctionFly. Here's what's coming:

- [x] Multi-language function execution (Go, Python, Node.js, Rust, and more)
- [x] Edge deployment with global distribution
- [x] ML Intelligence Layer (FlyMind)
- [x] MCP server integration
- [x] Built-in monitoring with Prometheus & Grafana
- [ ] Kubernetes operator and Helm chart
- [ ] Multi-region failover and HA
- [ ] Function versioning and canary deployments
- [ ] Advanced RBAC and org-level policies
- [ ] Expanded SDK support (Java, .NET, PHP)
- [ ] Managed cloud offering (FunctionFly Cloud)

See our public [project board](https://github.com/orgs/functionfly/projects) for current progress.

## Backers & Sponsors

FunctionFly is independently developed. If you find it valuable, consider:

- ⭐ Starring the repo to show your support
- 🐛 Reporting issues and contributing PRs
- 💬 Joining our [Discord](https://discord.com/invite/cSTsz3WjpD) community

## Contact

- **Website**: [functionfly.com](https://functionfly.com)
- **Discord**: [discord.com/invite/cSTsz3WjpD](https://discord.com/invite/cSTsz3WjpD)
- **Email**: support@functionfly.com
- **Docs**: [docs.functionfly.com](https://docs.functionfly.com)

---

*© 2026 FunctionFly Inc. — All rights reserved.*
