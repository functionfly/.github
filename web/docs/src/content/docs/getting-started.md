---
title: Getting Started
description: Quick start guide for FunctionFly™
---

# Getting Started with FunctionFly™

Welcome to FunctionFly™, a production-ready serverless function platform built with Go.

## Quick Start

```bash
# Install the ffly CLI
curl -fsSL https://raw.githubusercontent.com/functionfly/functionfly/main/scripts/install.sh | bash

# Login
ffly login

# Initialize a function
ffly init my-function

# Run locally
ffly dev

# Publish to registry
ffly publish
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
ffly login --help
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
ffly init hello-world
cd hello-world
ffly dev  # Test locally
ffly publish  # Deploy to production
```

## Next Steps

- [Functions Guide](/docs/functions) - Learn about writing functions
- [Deployment](/docs/deployment) - Deploy to production
- [API Reference](/docs/api) - Full API documentation
- [Examples](/docs/examples) - Sample functions
