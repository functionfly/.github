---
title: Supported Providers
description: Learn about all cloud providers supported by FunctionFly and their specific configuration options.
pubDate: 2024-01-16
updatedDate: 2024-01-16
heroImage: /docs/providers-hero.png
category: Reference
tags: [providers, aws, cloudflare, vercel, configuration]
draft: false
---

# Supported Providers

FunctionFly supports multiple serverless platforms to give you maximum flexibility and reliability. Here's a complete guide to configuring each provider.

## AWS Lambda

### Configuration

```javascript
{
  name: 'aws-lambda',
  region: 'us-east-1',
  accessKeyId: process.env.AWS_ACCESS_KEY_ID,
  secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
  functions: [
    {
      name: 'my-function',
      arn: 'arn:aws:lambda:us-east-1:123456789012:function:my-function',
      version: '$LATEST'
    }
  ]
}
```

### Features

- Automatic scaling
- Multiple runtime support
- VPC integration
- X-Ray tracing

## Cloudflare Workers

### Configuration

```javascript
{
  name: 'cloudflare-workers',
  accountId: process.env.CLOUDFLARE_ACCOUNT_ID,
  apiToken: process.env.CLOUDFLARE_API_TOKEN,
  functions: [
    {
      name: 'my-worker',
      script: 'my-worker-script',
      routes: ['api.example.com/*']
    }
  ]
}
```

### Features

- Global edge network
- WebAssembly support
- Built-in security features
- Real-time logs

## Vercel Functions

### Configuration

```javascript
{
  name: 'vercel',
  token: process.env.VERCEL_TOKEN,
  teamId: process.env.VERCEL_TEAM_ID,
  functions: [
    {
      name: 'my-api',
      url: 'https://my-app.vercel.app/api/my-function'
    }
  ]
}
```

### Features

- Git integration
- Preview deployments
- Analytics included
- Serverless function logs

## Provider Comparison

| Feature | AWS Lambda | Cloudflare Workers | Vercel |
|---------|------------|-------------------|--------|
| Cold Starts | Yes | Minimal | Minimal |
| Global CDN | Via CloudFront | Built-in | Built-in |
| Pricing Model | Per request | Per request | Per request |
| Runtime Limits | 15min | 30s | 10s |
| Languages | Many | JavaScript | JavaScript/Node |

## Best Practices

### Multi-Provider Setup

For maximum reliability, configure at least two providers:

```javascript
export default {
  name: 'my-app',
  providers: [
    {
      name: 'aws-lambda',
      region: 'us-east-1',
      weight: 60, // 60% of traffic
      functions: [...]
    },
    {
      name: 'cloudflare-workers',
      weight: 40, // 40% of traffic
      functions: [...]
    }
  ]
};
```

### Health Checks

Configure appropriate health check endpoints for each provider to ensure FunctionFly can detect failures quickly.

### Monitoring

Enable detailed logging and monitoring for each provider to get insights into performance and reliability.