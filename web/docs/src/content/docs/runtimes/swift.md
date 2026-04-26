---
title: Swift via WASM Runtime
description: Swift WebAssembly runtime environment for FunctionFly functions
---

# Swift via WASM Runtime

FunctionFly's Swift runtime compiles your Swift code to WebAssembly using SwiftWasm for type-safe, sandboxed execution.

## Overview

Swift functions are compiled to WebAssembly using the SwiftWasm toolchain. This provides:

- **Type safety** - Full Swift type system with optionals and error handling
- **Fast cold starts** - WASM modules initialize quickly
- **Sandboxed** - Memory-safe, isolated execution
- **Protocol-oriented** - Use Swift's protocol-oriented design
- **Value types** - Efficient struct-based data modeling

## Supported Toolchains

| Toolchain | Status | Notes |
|-----------|--------|-------|
| SwiftWasm | Supported | Recommended; direct compilation |
| carton | Supported | Build tool for SwiftWasm |

## Project Structure

```
my-function/
├── Package.swift           # Swift package manifest
├── Sources/
│   └── MyFunction/
│       └── main.swift      # Entry point
└── functionfly.jsonc       # Function config
```

## Package.swift

```swift
// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "my-function",
    platforms: [.macOS(.v12)],
    products: [
        .executable(name: "my-function", targets: ["MyFunction"])
    ],
    dependencies: [
        .package(url: "https://github.com/functionfly/sdk-swift", from: "1.0.0")
    ],
    targets: [
        .executableTarget(
            name: "My-function",
            dependencies: ["FunctionFly"]
        )
    ]
)
```

## Function Structure

A Swift function implements the `FunctionFlyFunction` protocol:

```swift
import FunctionFly

struct MyFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        return #"{"message": "Hello from Swift!"}"#
    }
}

FunctionFly.run(MyFunction())
```

## Example Functions

### HTTP API Handler

```swift
import FunctionFly

struct UserAPI: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        let data = input.data(using: .utf8)!
        let request = try JSONSerialization.jsonObject(with: data) as! [String: Any]
        let method = request["method"] as? String ?? "GET"

        switch method {
        case "GET":
            let users: [[String: Any]] = [
                ["id": "1", "name": "Alice", "email": "alice@example.com"],
                ["id": "2", "name": "Bob", "email": "bob@example.com"]
            ]
            let response: [String: Any] = ["status": 200, "body": ["users": users]]
            let jsonData = try JSONSerialization.data(withJSONObject: response)
            return String(data: jsonData, encoding: .utf8)!

        case "POST":
            let body = request["body"] as? [String: Any] ?? [:]
            let name = body["name"] as? String ?? ""
            let email = body["email"] as? String ?? ""
            let user: [String: Any] = ["id": "3", "name": name, "email": email]
            let response: [String: Any] = ["status": 201, "body": user]
            let jsonData = try JSONSerialization.data(withJSONObject: response)
            return String(data: jsonData, encoding: .utf8)!

        default:
            return #"{"status": 405, "body": {"error": "Method not allowed"}}"#
        }
    }
}

FunctionFly.run(UserAPI())
```

### Webhook Processor

```swift
import FunctionFly
import Crypto

struct WebhookHandler: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        let data = input.data(using: .utf8)!
        let request = try JSONSerialization.jsonObject(with: data) as! [String: Any]
        let headers = request["headers"] as? [String: String] ?? [:]
        let secret = FunctionFly.getEnv("WEBHOOK_SECRET")
        let signature = headers["x-signature"] ?? ""

        guard verifySignature(body: request["body"], signature: signature, secret: secret) else {
            return #"{"status": 401, "body": {"error": "Invalid signature"}}"#
        }

        FunctionFly.log("Webhook received")
        return #"{"status": 200, "body": {"received": true}}"#
    }

    func verifySignature(body: Any?, signature: String, secret: String) -> Bool {
        guard signature.hasPrefix("sha256=") else { return false }
        // In production, compute HMAC-SHA256 and compare
        return true
    }
}

FunctionFly.run(WebhookHandler())
```

### Data Transformation

```swift
import FunctionFly

struct DataTransformer: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        let data = input.data(using: .utf8)!
        let request = try JSONSerialization.jsonObject(with: data) as! [String: Any]
        let body = request["body"] as? [String: Any] ?? [:]

        let firstName = body["first_name"] as? String ?? ""
        let lastName = body["last_name"] as? String ?? ""
        let name = "\(firstName) \(lastName)".trimmingCharacters(in: .whitespaces)
        let email = (body["email"] as? String ?? "").lowercased()

        let transformed: [String: Any] = [
            "id": body["id"] ?? "",
            "name": name,
            "email": email,
            "timestamp": body["created_at"] ?? ""
        ]

        let response: [String: Any] = ["status": 200, "body": transformed]
        let jsonData = try JSONSerialization.data(withJSONObject: response)
        return String(data: jsonData, encoding: .utf8)!
    }
}

FunctionFly.run(DataTransformer())
```

## Environment Variables

Access environment variables using `FunctionFly.getEnv`:

```swift
let apiKey = FunctionFly.getEnv("API_KEY")
let dbUrl = FunctionFly.getEnv("DATABASE_URL")
let debug = FunctionFly.getEnv("DEBUG") == "true"
```

## Error Handling

```swift
struct SafeFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        do {
            let data = input.data(using: .utf8)!
            let request = try JSONSerialization.jsonObject(with: data) as! [String: Any]
            let result = try processRequest(request)
            let jsonData = try JSONSerialization.data(withJSONObject: ["status": 200, "body": result])
            return String(data: jsonData, encoding: .utf8)!
        } catch let error as DecodingError {
            return #"{"status": 400, "body": {"error": "Invalid input: \#(error.localizedDescription)"}}"#
        } catch {
            FunctionFly.log("Error: \(error)")
            return #"{"status": 500, "body": {"error": "\#(error.localizedDescription)"}}"#
        }
    }
}
```

## functionfly.jsonc Configuration

```jsonc
{
  "name": "my-swift-function",
  "runtime": "swift-wasm",
  "wasm": {
    "entrypoint": "main",
    "wasi": true
  },
  "limits": {
    "timeout": 30,
    "memory": 256
  }
}
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 1024 MB |
| CPU | 1 vCPU | 4 vCPU |

Swift/WASM functions may use more memory than C/Rust due to the Swift runtime.

## Cold Start

Swift/WASM functions have moderate cold starts:
- First invocation after deployment: ~40-120ms
- Subsequent invocations: ~3-10ms

## Best Practices

1. **Use structs over classes** - Value types are more efficient in WASM
2. **Handle errors with `do/catch`** - Always handle potential failures
3. **Use `JSONSerialization`** - For JSON parsing (Foundation is available)
4. **Prefer `FunctionFlyContext`** - Use context for host function access
5. **Keep dependencies minimal** - Each dependency adds to binary size
6. **Return valid JSON** - Always return a JSON string from `handle`
7. **Use optionals safely** - Leverage Swift's optional chaining
