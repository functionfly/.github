---
title: Ruby Runtime
description: Ruby runtime environment for FunctionFly functions
---

# Ruby Runtime

FunctionFly's Ruby runtime executes your Ruby code inside an embedded mruby interpreter compiled to WebAssembly.

## Overview

Ruby functions run inside a lightweight mruby interpreter embedded in WASM. This provides:

- **Familiar Ruby syntax** - Write functions in idiomatic Ruby
- **Fast cold starts** - mruby interpreter starts in milliseconds
- **Sandboxed** - Safe execution within WASM boundaries
- **No gems required** - mruby uses a minimal standard library

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| mruby 3.x | Supported | Primary runtime |
| Ruby 3.x (MRI) | Not supported | Use mruby-compatible syntax |

## Project Structure

```
my-function/
├── main.rb             # Entry point
├── functionfly.jsonc   # Function config
└── lib/                # Optional helper modules
    └── utils.rb
```

## Function Structure

A Ruby function subclasses `FunctionFly::Function`:

```ruby
class MyFunction < FunctionFly::Function
  def handle(input, ctx)
    '{"message": "Hello from Ruby!"}'
  end
end

FunctionFly.run(MyFunction)
```

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
      response(200, { "users" => users })
    when "POST"
      name = request.dig("body", "name") || ""
      email = request.dig("body", "email") || ""
      user = { "id" => "3", "name" => name, "email" => email }
      response(201, user)
    else
      response(405, { "error" => "Method not allowed" })
    end
  end

  private

  def response(status, body)
    { "status" => status, "body" => body }.to_json
  end
end

FunctionFly.run(UserAPI)
```

### Webhook Processor

```ruby
class WebhookHandler < FunctionFly::Function
  def handle(input, ctx)
    request = JSON.parse(input)
    secret = ENV["WEBHOOK_SECRET"] || ""
    signature = request.dig("headers", "x-signature") || ""

    unless verify_signature(request["body"], signature, secret)
      return { "status" => 401, "body" => { "error" => "Invalid signature" } }.to_json
    end

    # Process webhook event
    event = request["body"]
    FunctionFly.log("Received event: #{event["type"]}")

    { "status" => 200, "body" => { "received" => true } }.to_json
  end

  private

  def verify_signature(body, signature, secret)
    return false unless signature.start_with?("sha256=")

    expected = OpenSSL::HMAC.hexdigest("SHA256", secret, body.to_json)
    provided = signature[7..]

    # Constant-time comparison
    return false if expected.length != provided.length
    expected.bytes.zip(provided.bytes).each do |a, b|
      return false if a != b
    end
    true
  end
end

FunctionFly.run(WebhookHandler)
```

### Data Transformation

```ruby
class DataTransformer < FunctionFly::Function
  def handle(input, ctx)
    request = JSON.parse(input)
    data = request["body"]

    transformed = {
      "id" => data["id"],
      "name" => "#{data["first_name"]} #{data["last_name"]}".strip,
      "email" => data["email"].downcase,
      "timestamp" => data["created_at"]
    }

    { "status" => 200, "body" => transformed }.to_json
  end
end

FunctionFly.run(DataTransformer)
```

## SDK Methods

| Method | Description |
|--------|-------------|
| `FunctionFly.run(klass)` | Register and run a function class |
| `FunctionFly.log(message)` | Log a message to the execution log |
| `FunctionFly.get_env(key)` | Get an environment variable |

## Environment Variables

Access environment variables using `ENV`:

```ruby
class MyFunction < FunctionFly::Function
  def handle(input, ctx)
    api_key = ENV["API_KEY"] || ""
    db_url = ENV["DATABASE_URL"] || ""
    debug = ENV["DEBUG"] == "true"

    { "api_key_set" => !api_key.empty?, "debug" => debug }.to_json
  end
end
```

## File System

The `/tmp` directory is available for temporary storage:

```ruby
class FileHandler < FunctionFly::Function
  def handle(input, ctx)
    # Write to temp file
    File.write("/tmp/output.json", '{"result": "cached"}')

    # Read from temp file
    data = File.read("/tmp/output.json")

    data
  end
end
```

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

  private

  def process(request)
    # Business logic here
    { "status" => "ok" }
  end
end

FunctionFly.run(SafeFunction)
```

## Limitations

The mruby runtime has some limitations compared to MRI Ruby:

- **No gems** - Only mruby's built-in standard library is available
- **No threads** - Single-threaded execution only
- **No native extensions** - C extensions are not supported
- **Limited stdlib** - Subset of Ruby's standard library

## functionfly.jsonc Configuration

```jsonc
{
  "name": "my-ruby-function",
  "runtime": "ruby-wasm",
  "wasm": {
    "entrypoint": "main",
    "interpreter": "mruby"
  },
  "limits": {
    "timeout": 30,
    "memory": 128
  }
}
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 128 MB | 512 MB |
| CPU | 1 vCPU | 4 vCPU |

## Cold Start

Ruby functions have moderate cold starts due to interpreter initialization:
- First invocation after deployment: ~50-150ms
- Subsequent invocations: ~5-15ms

## Best Practices

1. **Use mruby-compatible syntax** - Avoid MRI-specific features
2. **Parse JSON carefully** - Always handle parse errors
3. **Keep functions small** - Minimize code loaded at startup
4. **Avoid deep nesting** - mruby has limited stack depth
5. **Use `FunctionFly.log`** - For debugging and monitoring
6. **Return JSON strings** - Always return valid JSON from `handle`
7. **Handle all errors** - Wrap business logic in error handlers
