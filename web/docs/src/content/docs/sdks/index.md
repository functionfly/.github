---
title: SDKs
description: Official FunctionFly SDKs for Python, JavaScript, Go, Rust, and more
---

## Official SDKs

FunctionFly provides official SDKs for all major languages. Each SDK wraps
the REST API with idiomatic types, error handling, and convenience methods.

| SDK | Language | Install |
|-----|----------|---------|
| [Python](/sdks/python/) | Python 3.11+ | `pip install functionfly` |
| [JavaScript / TypeScript](/sdks/javascript/) | Node.js 18+, Bun, Deno | `npm install @functionfly/sdk` |
| [Go](/sdks/go/) | Go 1.21+ | `go get github.com/functionfly/sdk-go` |
| [Rust](/sdks/rust/) | Rust (native + WASM) | `cargo add functionfly` |
| [C](/sdks/c/) | C / C++ | Header-only library |
| [Ruby](/sdks/ruby/) | Ruby 3+ | `gem install functionfly` |
| [Kotlin](/sdks/kotlin/) | Kotlin / JVM | `implementation("com.functionfly:sdk-kotlin")` |
| [Swift](/sdks/swift/) | Swift 5.9+ | `.package(url: "functionfly/sdk-swift")` |

## Common Patterns

All SDKs follow the same patterns:

### Authentication

```python
from functionfly import Client

client = Client(api_key="ffp_v1_...")
```

### Execute a Function

```python
result = client.functions.execute("author/name", {"input": "data"})
print(result.body)
```

### List Functions

```python
functions = client.functions.list()
for fn in functions:
    print(fn.name)
```

### Manage API Keys

```python
keys = client.api_keys.list()
new_key = client.api_keys.create(name="ci-deploy", key_type="platform")
```

## Next Steps

- Pick an SDK from the list above for language-specific guides
- [REST API](/api/) — Full API reference
- [Authentication](/api/authentication/) — Auth methods
- [CLI](/cli/) — Command-line interface
