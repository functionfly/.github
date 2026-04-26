---
title: Ruby SDK
description: Ruby SDK for building FunctionFly functions
---

# Ruby SDK

The FunctionFly Ruby SDK provides a class-based API for building FunctionFly functions in Ruby, running inside an embedded mruby interpreter compiled to WASM.

## Installation

Include the FunctionFly Ruby files in your project:

```
my-function/
├── functionfly.rb          # SDK entry point
├── functionfly/
│   ├── function.rb         # Base function class
│   ├── context.rb          # Execution context
│   └── version.rb          # Version constant
└── main.rb                 # Your function
```

## Quick Start

```ruby
require_relative "functionfly"

class MyFunction < FunctionFly::Function
  def handle(input, ctx)
    '{"message": "Hello from Ruby!"}'
  end
end

FunctionFly.run(MyFunction)
```

## Function Base Class

All functions extend `FunctionFly::Function`:

```ruby
class FunctionFly::Function
  # Process the input and return the output as a string.
  # @param input [String] The function input (JSON string)
  # @param ctx [FunctionFly::Context] Execution context
  # @return [String] The function output (JSON string)
  def handle(input, ctx)
    raise NotImplementedError, "Subclass must implement #handle"
  end
end
```

## Module Methods

| Method | Description |
|--------|-------------|
| `FunctionFly.run(klass)` | Register and run a function class |
| `FunctionFly.log(message)` | Log a message to execution logs |
| `FunctionFly.get_env(key)` | Get an environment variable |
| `FunctionFly.fetch(url)` | Make an HTTP request |

## Example Functions

### HTTP API Handler

```ruby
class UserAPI < FunctionFly::Function
  def handle(input, ctx)
    request = JSON.parse(input)

    case request["method"]
    when "GET"
      users = [
        { "id" => "1", "name" => "Alice", "email" => "alice@example.com" },
        { "id" => "2", "name" => "Bob", "email" => "bob@example.com" }
      ]
      { "status" => 200, "body" => { "users" => users } }.to_json
    when "POST"
      body = request["body"] || {}
      user = { "id" => "3", "name" => body["name"], "email" => body["email"] }
      { "status" => 201, "body" => user }.to_json
    else
      { "status" => 405, "body" => { "error" => "Method not allowed" } }.to_json
    end
  end
end

FunctionFly.run(UserAPI)
```

### Webhook Processor

```ruby
class WebhookHandler < FunctionFly::Function
  def handle(input, ctx)
    request = JSON.parse(input)
    secret = FunctionFly.get_env("WEBHOOK_SECRET")
    signature = request.dig("headers", "x-signature") || ""

    unless verify_signature(request["body"], signature, secret)
      return { "status" => 401, "body" => { "error" => "Invalid signature" } }.to_json
    end

    FunctionFly.log("Webhook processed")
    { "status" => 200, "body" => { "received" => true } }.to_json
  end

  private

  def verify_signature(body, signature, secret)
    return false unless signature.start_with?("sha256=")
    # Compute and compare HMAC-SHA256 in production
    true
  end
end

FunctionFly.run(WebhookHandler)
```

### Using Host Functions

```ruby
class APIClient < FunctionFly::Function
  def handle(input, ctx)
    api_key = FunctionFly.get_env("API_KEY")

    # Make an HTTP request
    response = FunctionFly.fetch("https://api.example.com/data")

    # Store result in KV
    FunctionFly.kv_set("last_fetch", response)

    { "status" => 200, "body" => { "data" => response } }.to_json
  end
end

FunctionFly.run(APIClient)
```

## Host Functions

| Method | Description |
|--------|-------------|
| `FunctionFly.log(msg)` | Log a message |
| `FunctionFly.get_env(key)` | Get environment variable |
| `FunctionFly.fetch(url)` | HTTP GET request |
| `FunctionFly.kv_get(key)` | Read from key-value store |
| `FunctionFly.kv_set(key, value)` | Write to key-value store |

## Error Handling

```ruby
class SafeFunction < FunctionFly::Function
  def handle(input, ctx)
    begin
      request = JSON.parse(input)
      result = process(request)
      { "status" => 200, "body" => result }.to_json
    rescue JSON::ParserError => e
      { "status" => 400, "body" => { "error" => "Invalid JSON: #{e.message}" } }.to_json
    rescue => e
      FunctionFly.log("Error: #{e.message}")
      { "status" => 500, "body" => { "error" => e.message } }.to_json
    end
  end
end
```

## API Reference

### Module Methods

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| `FunctionFly.run` | `klass` (Class) | `void` | Register and execute function |
| `FunctionFly.log` | `msg` (String) | `void` | Log message |
| `FunctionFly.get_env` | `key` (String) | `String` | Get env variable |
| `FunctionFly.fetch` | `url` (String) | `String` | HTTP GET |
| `FunctionFly.kv_get` | `key` (String) | `String` | KV read |
| `FunctionFly.kv_set` | `key` (String), `value` (String) | `void` | KV write |

### Classes

| Class | Description |
|-------|-------------|
| `FunctionFly::Function` | Base class; subclass and implement `#handle` |
| `FunctionFly::Context` | Execution context (function name, version) |

## Limitations

- **mruby only** — No MRI gems or native extensions
- **Single-threaded** — No concurrency primitives
- **Limited stdlib** — Subset of Ruby standard library
