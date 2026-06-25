# Migration Guide

Welcome to FunctionFly! This guide helps you migrate from other serverless platforms.

## Comparing Platforms

| Feature | FunctionFly | AWS Lambda | Vercel Functions | Cloudflare Workers |
|---------|-------------|------------|------------------|-------------------|
| Cold Start | ~50ms | ~100-500ms | ~50-200ms | ~5-50ms |
| Max Timeout | 300s | 900s | 10s ( Hobby) / 60s (Pro) | 50ms-30s |
| Max Memory | 1GB | 10GB | 3008MB | 512MB (unbound) |
| Runtime Support | 10+ languages | 6 languages | Node.js, Python | V8 Isolates |
| Trust Protocol | Built-in | None | None | None |
| Function DNA | Built-in | None | None | None |
| Registry | Built-in | ECR | None | Workers |
| Free Tier | Generous | 1M requests | 100k/day (Hobby) | 100k/day |

## Migrating from AWS Lambda

### Key Differences

- **No IAM roles needed**: FunctionFly uses API keys and OAuth
- **Built-in registry**: No ECR setup required
- **Trust certificates**: Automatic execution verification
- **Function DNA**: Automatic function analysis and scoring

### Step 1: Export Your Lambda Functions

```bash
# List your Lambda functions
aws lambda list-functions --query 'Functions[*].[FunctionName,Runtime,MemorySize,Timeout]'

# Download function code (one at a time)
aws lambda get-function --function-name my-function --query 'Code.Location'
```

### Step 2: Create FunctionFly Functions

```bash
# Install FunctionFly CLI
npm install -g @functionfly/cli

# Login
ffly login

# Create a new function
ffly init my-function --runtime node20

# Deploy
ffly deploy
```

### Step 3: Update Your Code

FunctionFly uses familiar patterns:

```javascript
// AWS Lambda
exports.handler = async (event) => {
  return { statusCode: 200, body: JSON.stringify({ message: 'Hello' }) };
};

// FunctionFly
export default async function handler(context) {
  return { statusCode: 200, body: JSON.stringify({ message: 'Hello' }) };
}
```

### Step 4: Migrate Environment Variables

```bash
# Export from Lambda
aws lambda get-function-configuration --function-name my-function \
  --query 'Environment.Variables' --output json > env.json

# Import to FunctionFly
ffly env set --env-file env.json
```

## Migrating from Vercel Functions

### Key Differences

- **No framework required**: FunctionFly works with any HTTP handler
- **Longer timeouts**: Up to 300s vs Vercel's 60s
- **Built-in state**: StateFabric for persistent data
- **Trust verification**: Cryptographic execution certificates

### Step 1: Export from Vercel

```bash
# Download your project
vercel download

# Export environment variables
vercel env pull .env.vercel
```

### Step 2: Deploy to FunctionFly

```bash
# Create project
ffly init my-project

# Copy files
cp -r my-vercel-project/* ./my-project/

# Deploy
ffly deploy --prod
```

### Code Migration

```javascript
// Vercel (Next.js API route)
export default function handler(req, res) {
  res.status(200).json({ message: 'Hello' });
}

// FunctionFly
export default async function handler(context) {
  return { statusCode: 200, body: JSON.stringify({ message: 'Hello' }) };
}
```

## Migrating from Cloudflare Workers

### Key Differences

- **Full runtime support**: Python, Go, Rust, and more (not just V8 JS)
- **Longer timeouts**: Up to 300s vs Workers' 30s
- **Persistent storage**: Built-in StateFabric
- **Traditional HTTP**: REST instead of Request/Response

### Step 1: Export Workers

```bash
# Download Worker code from Cloudflare Dashboard
# Or use wrangler to download
wrangler whoami  # verify authentication
```

### Step 2: Deploy to FunctionFly

```bash
# Create function
ffly init my-worker --runtime deno

# Deploy
ffly deploy
```

### Code Migration

```javascript
// Cloudflare Worker
export default {
  async fetch(request) {
    return new Response('Hello', { status: 200 });
  }
};

// FunctionFly
export default async function handler(context) {
  return { statusCode: 200, body: 'Hello' };
}
```

## Migration Toolkit

### Automatic Migration Assistant

```bash
# Run the migration assistant
ffly migrate --from lambda --project my-lambda-functions/

# Preview changes before applying
ffly migrate --from lambda --project my-lambda-functions/ --dry-run
```

### Supported Migrations

| Source | Status | Notes |
|--------|--------|-------|
| AWS Lambda | Beta | Supports Node.js, Python |
| Vercel Functions | Beta | Supports Next.js API routes |
| Cloudflare Workers | Beta | Supports Workers (V8) |
| Netlify Functions | Planned | Coming soon |
| Azure Functions | Planned | Coming soon |

## Post-Migration

### Verify Your Deployment

```bash
# Test your function
ffly invoke my-function --data '{"key": "value"}'

# Check execution receipts
ffly receipts list --function my-function
```

### Enable Trust Verification

```bash
# Enable execution certificates
ffly function update my-function --enable-verification

# Verify an execution
ffly verify --execution-id exec_abc123
```

### Set Up Monitoring

```bash
# View analytics
ffly analytics view --function my-function

# Set up alerts
ffly alerts create --function my-function --error-rate-threshold 5
```

## Need Help?

- **Discord**: Join our [community server](https://discord.gg/functionfly)
- **Support**: Email support@functionfly.com
- **Consulting**: Need hands-on help? [Contact us](https://functionfly.com/consulting)
