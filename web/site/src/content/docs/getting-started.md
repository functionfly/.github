---
title: Getting Started with FunctionFly
description: Learn how to set up FunctionFly for your serverless applications in under 10 minutes.
pubDate: 2024-01-15
updatedDate: 2024-01-15
heroImage: /docs/getting-started-hero.png
category: Getting Started
tags: [setup, tutorial, quickstart]
draft: false
---

# Getting Started with FunctionFly

Welcome to FunctionFly! This guide will help you get up and running with multi-cloud serverless failover in under 10 minutes.

## Prerequisites

Before you begin, make sure you have:

- AWS Lambda functions deployed
- Cloudflare Workers (optional)
- Vercel functions (optional)
- A FunctionFly account

## Step 1: Install the CLI

```bash
npm install -g functionfly-cli
# or
curl -fsSL https://cli.functionfly.com/install.sh | sh
```

## Step 2: Authenticate

```bash
ffly login
```

This will open your browser for authentication.

## Step 3: Create Your First Project

```bash
ffly projects create my-app
cd my-app
```

## Step 4: Configure Providers

Create a `functionfly.config.js` file:

```javascript
export default {
  name: 'my-app',
  providers: [
    {
      name: 'aws-lambda',
      region: 'us-east-1',
      functions: [
        {
          name: 'my-function',
          arn: 'arn:aws:lambda:us-east-1:123456789012:function:my-function'
        }
      ]
    },
    {
      name: 'cloudflare-workers',
      accountId: 'your-account-id',
      functions: [
        {
          name: 'my-worker',
          script: 'my-worker-script'
        }
      ]
    }
  ],
  failover: {
    strategy: 'latency-based',
    healthChecks: {
      interval: 30,
      timeout: 10,
      unhealthyThreshold: 3,
      healthyThreshold: 2
    }
  }
};
```

## Step 5: Deploy

```bash
ffly deploy
```

That's it! Your functions are now protected with automatic failover across multiple cloud providers.

## Next Steps

- [Configure monitoring](/docs/monitoring)
- [Set up alerts](/docs/alerts)
- [Advanced routing rules](/docs/routing)

## Need Help?

Join our [Discord community](https://discord.gg/functionfly) or check out our [GitHub issues](https://github.com/functionfly/functionfly/issues).