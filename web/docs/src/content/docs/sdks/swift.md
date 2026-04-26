---
title: Swift SDK
description: Swift SDK for building FunctionFly functions
---

# Swift SDK

The FunctionFly Swift SDK provides a protocol-based API for building FunctionFly functions in Swift, compiled to WebAssembly via SwiftWasm.

## Installation

Add the FunctionFly SDK to your `Package.swift`:

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
            name: "MyFunction",
            dependencies: ["FunctionFly"]
        )
    ]
)
```

## Quick Start

```swift
import FunctionFly

struct MyFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        return #"{"message": "Hello from Swift!"}"#
    }
}

FunctionFly.run(MyFunction())
```

## FunctionFlyFunction Protocol

All functions implement the `FunctionFlyFunction` protocol:

```swift
public protocol FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String
}
```

## Module API

| Method | Description |
|--------|-------------|
| `FunctionFly.run(function)` | Register and run a function |
| `FunctionFly.log(message)` | Log a message to execution logs |
| `FunctionFly.getEnv(key)` | Get an environment variable |

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
            let user: [String: Any] = [
                "id": "3",
                "name": body["name"] as? String ?? "",
                "email": body["email"] as? String ?? ""
            ]
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

### Using Environment Variables

```swift
import FunctionFly

struct ConfiguredFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        let apiKey = FunctionFly.getEnv("API_KEY")
        let debug = FunctionFly.getEnv("DEBUG") == "true"

        FunctionFly.log("Processing with debug=\(debug)")

        return #"{"api_key_set": \#(apiKey.isEmpty ? "false" : "true"), "debug": \#(debug)}"#
    }
}

FunctionFly.run(ConfiguredFunction())
```

## JSON Handling

The SDK uses Foundation's `JSONSerialization` for JSON:

```swift
// Parse
let data = input.data(using: .utf8)!
let json = try JSONSerialization.jsonObject(with: data) as! [String: Any]

// Build
let response: [String: Any] = [
    "status": 200,
    "body": ["key": "value", "count": 42]
]
let jsonData = try JSONSerialization.data(withJSONObject: response)
let jsonString = String(data: jsonData, encoding: .utf8)!
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
        } catch {
            FunctionFly.log("Error: \(error)")
            return #"{"status": 500, "body": {"error": "\#(error.localizedDescription)"}}"#
        }
    }
}
```

## Building for WASM

```bash
# Using carton (recommended)
carton build

# Output location
# .build/wasm32-unknown-wasi/release/my-function.wasm
```

## API Reference

### Protocols

| Protocol | Method | Signature |
|----------|--------|-----------|
| `FunctionFlyFunction` | `handle` | `func handle(input: String, context: FunctionFlyContext) throws -> String` |

### Module Functions

| Function | Parameters | Returns | Description |
|----------|-----------|---------|-------------|
| `FunctionFly.run` | `function: FunctionFlyFunction` | `void` | Register and execute |
| `FunctionFly.log` | `message: String` | `void` | Log message |
| `FunctionFly.getEnv` | `key: String` | `String` | Get env variable |

### Types

| Type | Description |
|------|-------------|
| `FunctionFlyContext` | Execution context (function name, version) |
