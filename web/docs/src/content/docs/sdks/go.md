---
title: Go SDK
description: Go SDK for FunctionFly
---

# Go SDK

The FunctionFly Go SDK provides a convenient way to interact with the FunctionFly API from Go applications.

## Installation

```bash
go get github.com/functionfly/sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    functionfly "github.com/functionfly/sdk-go"
)

func main() {
    // Initialize client
    client, err := functionfly.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List functions
    functions, err := client.Functions.List(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Invoke a function
    result, err := client.Functions.Invoke(ctx, "my-function", map[string]any{
        "name": "World",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result) // map[message:Hello, World!]
}
```

## Authentication

```go
import functionfly "github.com/functionfly/sdk-go"

// Using API key directly
client, err := functionfly.NewClient("ffly_...")

// Using environment variable
client, err := functionfly.NewClientFromEnv()
// Reads FFLY_API_KEY from environment
```

## Managing Functions

### Create a Function

```go
// Create from directory
function, err := client.Functions.Create(ctx, functionfly.CreateOptions{
    Name:    "my-api",
    Runtime: "go",
    Directory: "./my-function",
})

// Deploy
err := client.Functions.Deploy(ctx, function.ID)
```

### List Functions

```go
// List all functions
functions, err := client.Functions.List(ctx, nil)

// List with pagination
functions, err := client.Functions.List(ctx, &functionfly.ListOptions{
    Limit:  10,
    Offset: 0,
})

// Filter by runtime
functions, err := client.Functions.List(ctx, &functionfly.ListOptions{
    Runtime: "go",
})
```

### Get Function Details

```go
function, err := client.Functions.Get(ctx, "my-function")
if err != nil {
    log.Fatal(err)
}

fmt.Println(function.Name)
fmt.Println(function.Runtime)
fmt.Println(function.Version)
```

### Update a Function

```go
err := client.Functions.Update(ctx, "my-function", functionfly.UpdateOptions{
    Description: "Updated description",
    Environment: map[string]string{
        "DEBUG": "true",
    },
})
```

### Delete a Function

```go
err := client.Functions.Delete(ctx, "my-function")
```

## Invoking Functions

### Basic Invocation

```go
result, err := client.Functions.Invoke(ctx, "my-function", map[string]any{
    "key": "value",
})
```

### Typed Invocation

```go
type Request struct {
    Name string `json:"name"`
}

type Response struct {
    Message string `json:"message"`
}

req := Request{Name: "World"}
var resp Response

err := client.Functions.InvokeTyped(ctx, "my-function", &req, &resp)
fmt.Println(resp.Message) // Hello, World!
```

### With Custom Headers

```go
result, err := client.Functions.InvokeWithHeaders(
    ctx,
    "my-function",
    map[string]any{"key": "value"},
    map[string]string{
        "Authorization": "Bearer token",
        "X-Custom-Header": "value",
    },
)
```

## Environment Variables & Secrets

```go
// Set environment variables
err := client.Functions.SetEnv(ctx, "my-function", map[string]string{
    "API_URL": "https://api.example.com",
})

// Set secrets
err := client.Functions.SetSecrets(ctx, "my-function", map[string]string{
    "API_KEY": "secret-value",
})

// Get environment variables
env, err := client.Functions.GetEnv(ctx, "my-function")
```

## Monitoring

```go
// Get execution logs
logs, err := client.Functions.Logs(ctx, "my-function", &functionfly.LogOptions{
    Limit: 100,
})

// Get metrics
metrics, err := client.Functions.Metrics(ctx, "my-function")
fmt.Println(metrics.Invocations)
fmt.Println(metrics.Errors)
fmt.Println(metrics.AverageDuration)

// Check health
health, err := client.Functions.Health(ctx, "my-function")
fmt.Println(health.Status) // "healthy" or "unhealthy"
```

## Error Handling

```go
import (
    "errors"
    functionfly "github.com/functionfly/sdk-go"
)

result, err := client.Functions.Invoke(ctx, "nonexistent-function", nil)
if err != nil {
    var notFoundErr *functionfly.NotFoundError
    if errors.As(err, &notFoundErr) {
        fmt.Println("Function not found")
    }

    var authErr *functionfly.AuthenticationError
    if errors.As(err, &authErr) {
        fmt.Println("Invalid API key")
    }

    fmt.Printf("Error: %v\n", err)
}
```

## Configuration

```go
import (
    "time"
    functionfly "github.com/functionfly/sdk-go"
)

client, err := functionfly.NewClient(
    "ffly_...",
    functionfly.WithBaseURL("https://api.functionfly.com"),
    functionfly.WithTimeout(30*time.Second),
    functionfly.WithRetries(3),
)
```

## Advanced Usage

### Concurrent Invocations

```go
import "sync"

func batchInvoke(ctx context.Context, client *functionfly.Client) {
    functions := []string{"func1", "func2", "func3"}
    
    var wg sync.WaitGroup
    results := make([]any, len(functions))
    errors := make([]error, len(functions))

    for i, name := range functions {
        wg.Add(1)
        go func(index int, funcName string) {
            defer wg.Done()
            
            result, err := client.Functions.Invoke(ctx, funcName, map[string]any{
                "id": index,
            })
            results[index] = result
            errors[index] = err
        }(i, name)
    }

    wg.Wait()
    // Process results and errors
}
```

### With Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := client.Functions.Invoke(ctx, "my-function", nil)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Request timed out")
    }
}
```

### Streaming Results

```go
stream, err := client.Functions.InvokeStream(ctx, "streaming-function", nil)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for stream.Next() {
    var chunk map[string]any
    if err := stream.Decode(&chunk); err != nil {
        log.Fatal(err)
    }
    fmt.Println(chunk)
}
```

## Testing

### Mock Client

```go
import (
    "testing"
    functionfly "github.com/functionfly/sdk-go"
)

func TestMyFunction(t *testing.T) {
    // Create mock client
    mock := functionfly.NewMockClient()
    
    // Set up expectations
    mock.On("Functions.Invoke", "test-function", map[string]any{
        "name": "World",
    }).Return(map[string]any{
        "message": "Hello, World!",
    }, nil)

    // Use mock in your code
    result, err := mock.Functions.Invoke(ctx, "test-function", map[string]any{
        "name": "World",
    })
    
    if err != nil {
        t.Fatal(err)
    }
    
    if result["message"] != "Hello, World!" {
        t.Errorf("Unexpected result: %v", result)
    }
}
```

## API Reference

### Client Options

| Option | Type | Description |
|--------|------|-------------|
| `WithBaseURL(url string)` | Option | Set custom API base URL |
| `WithTimeout(timeout time.Duration)` | Option | Set request timeout |
| `WithRetries(n int)` | Option | Set number of retries |

### Types

```go
type Function struct {
    ID          string
    Name        string
    Runtime     string
    Version     string
    Description string
    Status      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type InvocationResult struct {
    Status     int
    Body       any
    Headers    map[string]string
    Duration   time.Duration
    ExecutedAt time.Time
}

type Metrics struct {
    Invocations      int64
    Errors           int64
    AverageDuration  time.Duration
    P95Duration      time.Duration
    P99Duration      time.Duration
}
```

### Errors

| Error Type | Description |
|------------|-------------|
| `NotFoundError` | Resource not found |
| `AuthenticationError` | Invalid or missing API key |
| `AuthorizationError` | Insufficient permissions |
| `RateLimitError` | Rate limit exceeded |
| `ValidationError` | Invalid request data |
