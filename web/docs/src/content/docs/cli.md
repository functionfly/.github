---
title: CLI Reference
description: The ff CLI has moved to its own repository
---

# CLI Reference

The `ff` CLI has moved to its own repository: **[functionfly/ff-cli](https://github.com/functionfly/ff-cli)**

For installation instructions and full CLI documentation, please visit the [ff-cli repository](https://github.com/functionfly/ff-cli).

## Quick Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.ps1 | iex

# Homebrew
brew tap functionfly/tap && brew install ff
```

## Quick Start

```bash
# Login to FunctionFly
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

For complete documentation, see [https://github.com/functionfly/ff-cli](https://github.com/functionfly/ff-cli)