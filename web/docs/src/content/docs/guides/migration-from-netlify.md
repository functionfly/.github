---
title: Migration from Netlify
description: Step-by-step guide to migrating your Netlify Functions to FunctionFly.
---

# Migration from Netlify

This guide walks you through migrating functions from Netlify to FunctionFly.

## Overview

| Aspect | Netlify Functions | FunctionFly |
|--------|-------------------|-------------|
| Timeout | 10s (free) / 26s (pro) | Up to 300s |
| Runtimes | Node.js, Go | 10+ languages |
| Build integration | Required | Optional |
| State | External | Built-in StateFabric |

## Prerequisites

```bash
# Install FunctionFly CLI
npm install -g @functionfly/cli

# Login
ffly login
```

## Step 1: Export from Netlify

```bash
# Download your site
# From dashboard: Site settings > Deploys > Download production build

# Export environment variables
# Site settings > Environment variables > Export .env file
```

## Step 2: Create FunctionFly Project

```bash
# Create project
ffly init my-netlify-functions

# Copy your function files
cp -r netlify/functions/* ./functions/
```

## Step 3: Update Handler Code

### Netlify Functions

```javascript
// Netlify Function
exports.handler = async (event, context) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello' })
  };
};

// FunctionFly
export default async function handler(context) {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'Hello' })
  };
}
```

### Netlify Build-less Functions (Beta)

```javascript
// Netlify (serverless function)
const axios = require('axios');

exports.handler = async (event) => {
  const response = await axios.get('https://api.example.com/data');
  return {
    statusCode: 200,
    body: JSON.stringify(response.data)
  };
};

// FunctionFly
import axios from 'axios';

export default async function handler(context) {
  const response = await axios.get('https://api.example.com/data');
  return {
    statusCode: 200,
    body: JSON.stringify(response.data)
  };
}
```

## Step 4: Migrate Environment Variables

```bash
# Import from Netlify export
ffly env set --env-file .env.netlify
```

## Step 5: Deploy

```bash
# Deploy
ffly deploy --prod
```

## Key Differences

| Netlify Concept | FunctionFly Equivalent |
|-----------------|------------------------|
| Functions | Functions |
| Build hooks | Triggers |
| Forms | Form handling |
| Identity | Auth (GBA) |
| Edge Handlers | Edge Runtime |
| Netlify Dev | `ffly dev` |

## Netlify-Specific Migrations

### Context and Client Context

```javascript
// Netlify
exports.handler = async (event, context) => {
  const { clientContext } = context;
  const user = clientContext.identity;
  return { statusCode: 200, body: JSON.stringify({ user }) };
};

// FunctionFly
export default async function handler(context) {
  const user = context.auth.user;
  return {
    statusCode: 200,
    body: JSON.stringify({ user })
  };
}
```

### Scheduled Functions (Cron)

```bash
# Netlify: netlify.toml
# [functions]
#   node_bundler = "eszip"
#
# [build.environment]
#   NODE_VERSION = "18"

# FunctionFly: ffly.yml
functions:
  my-cron:
    trigger: schedule
    schedule: "0 * * * *"
```

## Need Help?

- **Discord**: [community server](https://discord.gg/functionfly)
- **Support**: support@functionfly.com
