---
title: Getting Started
description: Quick start guide for FunctionFly™
---

# Getting Started with FunctionFly™

Welcome to FunctionFly™, a production-ready serverless function platform built with Go.

## Quick Start

```bash
# Install the ff CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.sh | bash

# Login
ff login

# Initialize a function
ff init my-function

# Run locally
ff dev

# Publish to registry
ff publish

# Or test your function instantly in the [Playground](/guides/playground/)
```

## Installation

### CLI Installation

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.sh | bash

# Homebrew (macOS)
brew tap functionfly/tap
brew install ff

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/functionfly/ff-cli/main/scripts/install.ps1 | iex
```

### Verify Installation

```bash
ff --version
ff login --help
```

## Your First Function

Create a new function and deploy it in minutes:

```javascript
// main.py - Python function
export default async function(req) {
  return {
    status: 200,
    body: `Hello, ${req.body.name || 'World'}!`
  };
}
```

```bash
ff init hello-world
cd hello-world
ff dev  # Test locally
ff publish  # Deploy to production
```

## Next Steps

- [Functions Guide](/docs/functions) - Learn about writing functions
- [Deployment](/docs/deployment) - Deploy to production
- [API Reference](/docs/api) - Full API documentation
- [Function Playground](/guides/playground/) - Test functions interactively
- [Examples](/docs/examples) - Sample functions
