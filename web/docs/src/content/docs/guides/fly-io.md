---
title: Fly.io
description: Deploy FunctionFly functions to Fly.io
---

# Fly.io

Deploy your FunctionFly functions to [Fly.io](https://fly.io/) for distributed applications with edge nodes close to your users.

## Features

- **Distributed edge** - 30+ regions worldwide
- **Anycast networking** - Automatic routing to nearest instance
- **Persistent storage** - Fly Volumes for data
- **Custom scaling** - Scale to zero and back

## Prerequisites

- Fly.io account
- Fly CLI installed: `npm i -g @flydotio/docker`

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "fly",
  "provider_config": {
    "app_name": "my-functionfly-app",
    "primary_region": "ord",
    "regions": ["ord", "iad", "lax", "sfo", "cdg"]
  }
}
```

## Deployment

```bash
# Deploy to Fly.io
ffly deploy --provider fly

# Deploy specific region
ffly deploy --provider fly --region ord
```

## Environment Variables

```bash
# Set Fly.io tokens
ffly env set FLY_API_TOKEN=your_token --provider fly
ffly env set FLY_APP_NAME=my-app --provider fly
```

## Fly Volumes

For persistent storage, FunctionFly uses Fly Volumes:

```javascript
// Volume mount at /vol/data
const fs = require('fs');
const data = fs.readFileSync('/vol/data/file.txt');
```

## Limitations

- Maximum 256MB memory per instance
- Maximum 120 second timeout
- No free tier for production
- Requires Docker for custom runtimes