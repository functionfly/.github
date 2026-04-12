---
title: Getting Started
description: Quick start guide for FunctionFly™
---

# Getting Started with FunctionFly™

Welcome to FunctionFly™, a production-ready serverless function platform built with Go.

## Quick Start

```bash
# Install the fly CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/functionfly/main/scripts/install.sh | bash

# Login
fly login

# Initialize a function
fly init my-function

# Run locally
fly dev

# Publish to registry
fly publish
```

## Installation

### CLI Installation

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/functionfly/functionfly/main/scripts/install.sh | bash

# Homebrew (macOS)
brew tap functionfly/tap
brew install ffly

# From source
go build -o bin/fly ./cmd/fly
```

### Verify Installation

```bash
fly --version
fly login --help
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
fly init hello-world
cd hello-world
fly dev  # Test locally
fly publish  # Deploy to production
```

## Next Steps

- [Functions Guide](/docs/functions) - Learn about writing functions
- [Deployment](/docs/deployment) - Deploy to production
- [API Reference](/docs/api) - Full API documentation
- [Examples](/docs/examples) - Sample functions
