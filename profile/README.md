# FunctionFly

<div align="center">

![FunctionFly Logo](https://functionfly.com/logo.png)

**A production-ready serverless function platform built for high-performance execution at the edge.**

[![Discord](https://img.shields.io/discord/123456789?label=Discord)](https://discord.gg/functionfly)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## What is FunctionFly?

FunctionFly™ is a comprehensive serverless platform that enables developers to deploy and run functions in multiple languages with automatic scaling, built-in monitoring, and a pay-per-use pricing model.

## Projects

| Repository | Description |
|------------|-------------|
| [fly](https://github.com/functionfly/fly) | Official CLI (`ffly`) for FunctionFly |
| [homebrew-tap](https://github.com/functionfly/homebrew-tap) | Homebrew tap for macOS/Linux installation |

## Getting Started

```bash
# Install the CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/fly/main/scripts/install.sh | bash

# Login and deploy your first function
ffly login
ffly deploy --path ./my-function
```

## Features

- **Multi-language Support** — Go, Python, Node.js, and more
- **Edge Execution** — Run functions close to your users globally
- **Automatic Scaling** — Scale from zero to millions of requests
- **Built-in Monitoring** — Real-time metrics with Prometheus & Grafana
- **Pay-per-use Pricing** — Pay only for what you use
- **Secure by Default** — Isolated execution with secrets management

## Resources

- [Documentation](https://docs.functionfly.com)
- [Discord Community](https://discord.gg/functionfly)
- [GitHub Issues](https://github.com/functionfly/fly/issues)

## License

MIT License © 2026 FunctionFly