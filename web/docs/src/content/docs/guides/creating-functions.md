---
title: Creating Functions
description: Learn how to write, test, and deploy serverless functions using Python, JavaScript, or Go.
sidebar:
  order: 1
---

This guide walks you through creating your first serverless function on FunctionFly.

## Prerequisites

Before you begin, make sure you have:

- A FunctionFly account (sign up at [functionfly.com](https://functionfly.com))
- The FunctionFly CLI installed (`go install github.com/functionfly/functionfly/cmd/fly@latest`)
- (Optional) One of the language-specific SDKs

## Step 1: Initialize a New Function

Create a new function using the CLI:

```bash
# Create a new function directory
ff init my-function

# Or specify a runtime
ff init my-function --runtime python
ff init my-function --runtime javascript
ff init my-function --runtime go
```

## Step 2: Write Your Function

### Python Example

```python
# main.py
import json

def handler(request):
    """Handle incoming requests."""
    name = request.get("name", "World")
    return {
        "statusCode": 200,
        "body": json.dumps({"message": f"Hello, {name}!"})
    }
```

### JavaScript Example

```javascript
// index.js
export default async function handler(request) {
  const name = request.name || 'World';
  return {
    statusCode: 200,
    body: JSON.stringify({ message: `Hello, ${name}!` })
  };
}
```

### Go Example

```go
// main.go
package main

import (
    "encoding/json"
    "net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        name = "World"
    }
    
    response := map[string]string{
        "message": "Hello, " + name + "!",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Step 3: Test Locally

Test your function locally before deploying:

```bash
# Start the local development server
ff dev

# In another terminal, test your function
curl http://localhost:8080 \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name": "FunctionFly"}'
```

## Step 4: Deploy to the Edge

Deploy your function to FunctionFly's global edge network:

```bash
# Deploy the function
ff deploy

# Get the deployment URL
ff info
```

## Step 5: Invoke Your Function

Once deployed, invoke your function via the API:

```bash
# Using curl
curl https://api.functionfly.com/v1/execute/<function-id> \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name": "FunctionFly"}'

# Or using the CLI
ffly invoke my-function --data '{"name": "FunctionFly"}'

## Testing with the Playground

After deploying, use the [Function Playground](/guides/playground/) to test your function interactively with different inputs, view execution history, and debug responses before going to production.

## Next Steps

- Learn about [deployment options](/deployment/)
- Set up [authentication](/guides/authentication/) for your function
- Explore the [registry](/guides/using-registry/) to find reusable functions
- Read about [secrets and vault](/guides/secrets-vault/) for sensitive data
- Test with the [Playground](/guides/playground/) before going live
