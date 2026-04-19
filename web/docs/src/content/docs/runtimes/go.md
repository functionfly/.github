---
title: Go Runtime
description: Go runtime environment for FunctionFly functions
---

# Go Runtime

FunctionFly's Go runtime compiles your Go code to binaries and executes them in a lightweight, fast environment.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Go 1.21 | Supported | Recommended |
| Go 1.22 | Supported | Latest |

## Function Structure

A Go function must implement the handler interface:

```go
package main

import (
    "context"
    "encoding/json"
)

// Request represents the incoming request
type Request struct {
    Body    json.RawMessage        `json:"body"`
    Headers map[string]string      `json:"headers"`
    Params  map[string]string      `json:"params"`
    Method  string                 `json:"method"`
    URL     string                 `json:"url"`
    Path    string                 `json:"path"`
}

// Response represents the function response
type Response struct {
    Status  int               `json:"status"`
    Body    interface{}       `json:"body"`
    Headers map[string]string `json:"headers,omitempty"`
}

// Handler processes the request and returns a response
func Handler(ctx context.Context, req Request) (Response, error) {
    return Response{
        Status: 200,
        Body:   map[string]string{"message": "Hello, World!"},
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
    }, nil
}
```

## Project Structure

```
my-function/
├── main.go           # Entry point
├── go.mod            # Go module
├── go.sum            # Dependencies
└── handlers/         # (optional)
    └── api.go
```

## go.mod

```go
module my-function

go 1.21

require (
    github.com/functionfly/sdk-go v0.1.0
)
```

## Request Type

```go
type Request struct {
    Body    json.RawMessage   `json:"body"`    // Raw JSON body
    Headers map[string]string `json:"headers"` // HTTP headers
    Params  map[string]string `json:"params"`  // Query parameters
    Method  string            `json:"method"`  // HTTP method
    URL     string            `json:"url"`     // Full URL
    Path    string            `json:"path"`    // URL path
}
```

## Response Type

```go
type Response struct {
    Status  int               `json:"status"`  // HTTP status code
    Body    interface{}       `json:"body"`    // Response body
    Headers map[string]string `json:"headers,omitempty"` // Response headers
}
```

## Example Functions

### HTTP API Handler

```go
package main

import (
    "context"
    "encoding/json"
)

type Request struct {
    Body    json.RawMessage   `json:"body"`
    Headers map[string]string `json:"headers"`
    Params  map[string]string `json:"params"`
    Method  string            `json:"method"`
    URL     string            `json:"url"`
    Path    string            `json:"path"`
}

type Response struct {
    Status  int               `json:"status"`
    Body    interface{}       `json:"body"`
    Headers map[string]string `json:"headers,omitempty"`
}

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func Handler(ctx context.Context, req Request) (Response, error) {
    switch req.Method {
    case "GET":
        users := []User{
            {ID: "1", Name: "Alice", Email: "alice@example.com"},
            {ID: "2", Name: "Bob", Email: "bob@example.com"},
        }
        return Response{
            Status: 200,
            Body:   map[string]interface{}{"users": users},
            Headers: map[string]string{"Content-Type": "application/json"},
        }, nil

    case "POST":
        var input struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        if err := json.Unmarshal(req.Body, &input); err != nil {
            return Response{
                Status: 400,
                Body:   map[string]string{"error": "Invalid JSON"},
            }, nil
        }

        user := User{
            ID:    "3",
            Name:  input.Name,
            Email: input.Email,
        }
        return Response{
            Status: 201,
            Body:   user,
            Headers: map[string]string{"Content-Type": "application/json"},
        }, nil

    default:
        return Response{
            Status: 405,
            Body:   map[string]string{"error": "Method not allowed"},
        }, nil
    }
}
```

### Webhook Processor

```go
package main

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "os"
    "strings"
)

type Request struct {
    Body    json.RawMessage   `json:"body"`
    Headers map[string]string `json:"headers"`
    Method  string            `json:"method"`
}

type Response struct {
    Status  int               `json:"status"`
    Body    interface{}       `json:"body"`
    Headers map[string]string `json:"headers,omitempty"`
}

func Handler(ctx context.Context, req Request) (Response, error) {
    // Verify signature
    signature := req.Headers["x-signature"]
    secret := os.Getenv("WEBHOOK_SECRET")

    if !verifySignature(req.Body, signature, secret) {
        return Response{
            Status: 401,
            Body:   map[string]string{"error": "Invalid signature"},
        }, nil
    }

    // Process webhook
    var event map[string]interface{}
    if err := json.Unmarshal(req.Body, &event); err != nil {
        return Response{
            Status: 400,
            Body:   map[string]string{"error": "Invalid JSON"},
        }, nil
    }

    // Process event asynchronously
    go processEvent(ctx, event)

    return Response{
        Status: 200,
        Body:   map[string]interface{}{"received": true},
    }, nil
}

func verifySignature(body []byte, signature, secret string) bool {
    if !strings.HasPrefix(signature, "sha256=") {
        return false
    }

    expected := hmac.New(sha256.New, []byte(secret))
    expected.Write(body)
    expectedSig := hex.EncodeToString(expected.Sum(nil))

    return hmac.Equal(
        []byte(strings.TrimPrefix(signature, "sha256=")),
        []byte(expectedSig),
    )
}

func processEvent(ctx context.Context, event map[string]interface{}) {
    // Your event processing logic
}
```

### Data Transformation

```go
package main

import (
    "context"
    "encoding/json"
    "strings"
)

type Request struct {
    Body json.RawMessage `json:"body"`
}

type Response struct {
    Status  int               `json:"status"`
    Body    interface{}       `json:"body"`
    Headers map[string]string `json:"headers,omitempty"`
}

type InputData struct {
    ID        string `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

type OutputData struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    Timestamp string `json:"timestamp"`
}

func Handler(ctx context.Context, req Request) (Response, error) {
    var input InputData
    if err := json.Unmarshal(req.Body, &input); err != nil {
        return Response{
            Status: 400,
            Body:   map[string]string{"error": "Invalid JSON"},
        }, nil
    }

    output := OutputData{
        ID:        input.ID,
        Name:      strings.TrimSpace(input.FirstName + " " + input.LastName),
        Email:     strings.ToLower(input.Email),
        Timestamp: input.CreatedAt,
    }

    return Response{
        Status: 200,
        Body:   output,
        Headers: map[string]string{"Content-Type": "application/json"},
    }, nil
}
```

## Environment Variables

Access environment variables using `os.Getenv`:

```go
import "os"

func Handler(ctx context.Context, req Request) (Response, error) {
    apiKey := os.Getenv("API_KEY")
    dbURL := os.Getenv("DATABASE_URL")
    debug := os.Getenv("DEBUG") == "true"
    
    // Use environment variables
    _ = apiKey
    _ = dbURL
    _ = debug
    
    return Response{Status: 200, Body: map[string]string{"status": "ok"}}, nil
}
```

## File System

The `/tmp` directory is available for temporary file storage:

```go
import (
    "os"
    "path/filepath"
)

func writeToTemp(filename string, data []byte) error {
    tmpPath := filepath.Join("/tmp", filename)
    return os.WriteFile(tmpPath, data, 0644)
}

func readFromTemp(filename string) ([]byte, error) {
    tmpPath := filepath.Join("/tmp", filename)
    return os.ReadFile(tmpPath)
}
```

Note: Files in `/tmp` are ephemeral and may not persist between invocations.

## Error Handling

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
)

type Request struct {
    Body   json.RawMessage `json:"body"`
    Method string          `json:"method"`
}

type Response struct {
    Status  int               `json:"status"`
    Body    interface{}       `json:"body"`
    Headers map[string]string `json:"headers,omitempty"`
}

func Handler(ctx context.Context, req Request) (Response, error) {
    result, err := processRequest(ctx, req)
    if err != nil {
        // Log error
        fmt.Printf("Error: %v\n", err)
        
        return Response{
            Status: 500,
            Body:   map[string]string{"error": "Internal server error"},
            Headers: map[string]string{"Content-Type": "application/json"},
        }, nil
    }

    return Response{
        Status: 200,
        Body:   result,
        Headers: map[string]string{"Content-Type": "application/json"},
    }, nil
}

func processRequest(ctx context.Context, req Request) (interface{}, error) {
    // Your processing logic
    return map[string]string{"status": "ok"}, nil
}
```

## Context Usage

Use the context for cancellation and timeouts:

```go
import (
    "context"
    "net/http"
    "time"
)

func Handler(ctx context.Context, req Request) (Response, error) {
    // Create HTTP client with context
    client := &http.Client{Timeout: 10 * time.Second}
    
    httpReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    if err != nil {
        return Response{Status: 500, Body: map[string]string{"error": err.Error()}}, nil
    }
    
    resp, err := client.Do(httpReq)
    if err != nil {
        return Response{Status: 500, Body: map[string]string{"error": err.Error()}}, nil
    }
    defer resp.Body.Close()
    
    return Response{
        Status: 200,
        Body:   map[string]string{"status": "ok"},
    }, nil
}
```

## Concurrent Operations

```go
import (
    "context"
    "sync"
)

func Handler(ctx context.Context, req Request) (Response, error) {
    var wg sync.WaitGroup
    results := make([]interface{}, 3)
    
    wg.Add(3)
    
    go func() {
        defer wg.Done()
        results[0] = fetchDataA(ctx)
    }()
    
    go func() {
        defer wg.Done()
        results[1] = fetchDataB(ctx)
    }()
    
    go func() {
        defer wg.Done()
        results[2] = fetchDataC(ctx)
    }()
    
    wg.Wait()
    
    return Response{
        Status: 200,
        Body:   map[string]interface{}{"results": results},
    }, nil
}
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 2048 MB |
| CPU | 1 vCPU | 4 vCPU |

Configure in `functionfly.jsonc`:

```jsonc
{
  "runtime": "go",
  "limits": {
    "timeout": 60,
    "memory": 512
  }
}
```

## Cold Start

Go functions have fast cold starts:
- First invocation after deployment: ~50-100ms
- Subsequent invocations: ~1-5ms

## Best Practices

1. **Keep main package small** - separate logic into packages
2. **Use proper types** - define Request/Response structs
3. **Handle errors** - return appropriate HTTP status codes
4. **Use context** - for timeouts and cancellation
5. **Minimize dependencies** - smaller binary size
6. **Use environment variables** - for configuration
7. **Avoid global state** - between invocations
