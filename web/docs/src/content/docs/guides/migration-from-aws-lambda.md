---
title: Migration from AWS Lambda
description: Step-by-step guide to migrating your AWS Lambda functions to FunctionFly.
---

# Migration from AWS Lambda

This guide walks you through migrating functions from AWS Lambda to FunctionFly.

## Overview

FunctionFly provides a simpler alternative to AWS Lambda:

| Aspect | AWS Lambda | FunctionFly |
|--------|------------|-------------|
| Authentication | IAM roles | API keys / OAuth |
| Container registry | ECR required | Built-in registry |
| Cold start | 100-500ms | ~50ms |
| Verification | CloudWatch | Trust Protocol |

## Prerequisites

```bash
# Install FunctionFly CLI
npm install -g @functionfly/cli

# Login
ffly login
```

## Step 1: Export Lambda Functions

```bash
# List all functions
aws lambda list-functions \
  --query 'Functions[*].[FunctionName,Runtime,MemorySize,Timeout]'

# Get function configuration
aws lambda get-function-configuration \
  --function-name YOUR_FUNCTION_NAME

# Download function code
aws lambda get-function \
  --function-name YOUR_FUNCTION_NAME \
  --query 'Code.Location'
```

## Step 2: Create FunctionFly Functions

```bash
# Create new function
ffly init my-function --runtime node20

# Navigate to function
cd my-function
```

## Step 3: Update Handler Code

### Node.js

```javascript
// AWS Lambda
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

### Python

```python
# AWS Lambda
def lambda_handler(event, context):
    return {
        'statusCode': 200,
        'body': json.dumps({'message': 'Hello'})
    }

# FunctionFly
def handler(context):
    return {
        'statusCode': 200,
        'body': json.dumps({'message': 'Hello'})
    }
```

## Step 4: Migrate Environment Variables

```bash
# Export from Lambda
aws lambda get-function-configuration \
  --function-name my-function \
  --query 'Environment.Variables' \
  --output json > env.json

# Import to FunctionFly
ffly env set --env-file env.json
```

## Step 5: Migrate Layers

FunctionFly uses bundles instead of layers:

```bash
# Add a bundle
ffly bundle add @functionfly/nodejs-utils

# Or create your own bundle
ffly bundle create my-utils --runtime node20
```

## Step 6: Set Timeout and Memory

```bash
# Update function settings
ffly function update my-function \
  --timeout 60 \
  --memory 512
```

## Step 7: Deploy and Test

```bash
# Deploy
ffly deploy --prod

# Test
ffly invoke my-function --data '{"key": "value"}'
```

## Verify Execution

```bash
# View execution receipts
ffly receipts list --function my-function

# Enable verification
ffly function update my-function --enable-verification
```

## Differences from Lambda

| Lambda Concept | FunctionFly Equivalent |
|----------------|------------------------|
| IAM Role | API Key / OAuth |
| ECR | Built-in Registry |
| CloudWatch Logs | `ffly logs` / Dashboard |
| Lambda Layers | Bundles |
| Event sources | Native triggers |
| SAM/CloudFormation | `ffly.yml` |

## Need Help?

- **Discord**: Join our [community server](https://discord.gg/functionfly)
- **Support**: support@functionfly.com
