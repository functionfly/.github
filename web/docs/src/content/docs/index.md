---
title: FunctionFly Documentation
description: The fastest way to deploy serverless functions to the global edge.
hero:
  title: FunctionFly
  tagline: Deploy serverless functions to 35+ edge locations worldwide
  actions:
    - text: Get Started
      link: /getting-started/
      icon: right-arrow
      variant: primary
    - text: View on GitHub
      link: https://github.com/functionfly
      icon: github
      variant: secondary
    - text: Try the Playground
      link: /guides/playground/
      icon: open-book
      variant: secondary
---

import { Card, CardGrid } from '@astrojs/starlight/components';

## Key Features

<CardGrid stagger>
  <Card title="35+ Edge Locations" icon="cloud">
    Deploy to major cities across North America, Europe, Asia, and Australia
  </Card>
  <Card title="Multi-runtime Support" icon="open-book">
    Write functions in JavaScript, TypeScript, Python, or Go
  </Card>
  <Card title="Enterprise Security" icon="seti:lock">
    Built-in malware scanning, signature verification, and trust levels
  </Card>
  <Card title="Real-time Analytics" icon="chart">
    Monitor function performance with detailed metrics
  </Card>
</CardGrid>

## Quick Start

```bash
# Install the ffly CLI (Go 1.25+)
go install github.com/functionfly/functionfly/cmd/fly@latest

# Login to FunctionFly
ffly login

# Initialize a new function
ffly init my-function

# Deploy to the edge
ffly deploy
```

## Documentation

Browse our documentation sections:

- **Getting Started** - Quick start guides and installation
- **Core Concepts** - Functions, CLI, and deployment
- **Trust & Security** - Trust API and security protocols
