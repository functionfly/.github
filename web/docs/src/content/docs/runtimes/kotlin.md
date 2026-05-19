---
title: Kotlin/JVM Runtime
description: Kotlin/JVM runtime environment for FunctionFly functions
---

FunctionFly's Kotlin/JVM runtime provides a secure, production-ready environment for executing Kotlin code with WASM sandbox isolation and comprehensive security controls.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Kotlin 1.9+ | Supported | Recommended |
| Java 17+ | Supported | JVM runtime |
| Kotlin/WASM | Supported | Via WASM sandbox |

## Architecture

The Kotlin runtime is built in Rust with the following components:

- **WASM Sandbox** - Code isolation via wasmtime
- **Security Manager** - Package/class blocking, host restrictions
- **Resource Limits** - Memory, CPU, wall time enforcement
- **Metrics Collection** - Prometheus-compatible observability
- **NATS Integration** - Orchestrator communication

## Function Structure

A Kotlin function must define a `main` function:

```kotlin
fun main() {
    val handler = Function { input ->
        """{"message": "Hello from Kotlin!"}"""
    }
    FunctionFly.run(handler)
}

fun interface Function {
    fun handle(input: String): String
}
```

## Request/Response Format

### Request

```kotlin
data class Request(
    val body: String,           // JSON body
    val headers: Map<String, String>, // HTTP headers
    val method: String,         // GET, POST, etc.
    val path: String,           // URL path
    val params: Map<String, String> // Query params
)
```

### Response

```kotlin
data class Response(
    val status: Int,                    // HTTP status code
    val body: String,                  // Response body (JSON string)
    val headers: Map<String, String>?  // Optional headers
)
```

## Example Functions

### HTTP API Handler

```kotlin
import kotlinx.serialization.json.*

fun main() {
    val api = Function { input ->
        val request = parseRequest(input)
        val response = when (request.method) {
            "GET" -> handleGet(request)
            "POST" -> handlePost(request)
            else -> createError(405, "Method not allowed")
        }
        serializeResponse(response)
    }
    FunctionFly.run(api)
}

fun handleGet(req: Request): Response {
    val users = """
        [{"id": "1", "name": "Alice"}, {"id": "2", "name": "Bob"}]
    """.trimIndent()
    return Response(
        status = 200,
        body = """{"users": $users}"""
    )
}

fun handlePost(req: Request): Response {
    val body = parseJson(req.body)
    val name = body["name"]?.toString() ?: ""
    return Response(
        status = 201,
        body = """{"id": "3", "name": $name}"""
    )
}
```

### Webhook Processor

```kotlin
import java.security.MessageDigest

fun main() {
    val webhook = Function { input ->
        val req = parseRequest(input)
        val signature = req.headers["x-signature"] ?: ""
        val secret = FunctionFly.getenv("WEBHOOK_SECRET")

        if (!verifySignature(req.body, signature, secret)) {
            return@Function serializeResponse(Response(401, """{"error": "Invalid signature"}"""))
        }

        FunctionFly.log("Webhook received: ${req.body}")
        serializeResponse(Response(200, """{"received": true}"""))
    }
    FunctionFly.run(webhook)
}

fun verifySignature(body: String, signature: String, secret: String): Boolean {
    if (!signature.startsWith("sha256=")) return false
    val expected = hmacSHA256(body, secret)
    return signature.substring(7) == expected
}
```

### Data Transformation

```kotlin
import kotlinx.serialization.json.*

fun main() {
    val transformer = Function { input ->
        val req = parseRequest(input)
        val body = parseJson(req.body)

        val firstName = body["first_name"]?.toString() ?: ""
        val lastName = body["last_name"]?.toString() ?: ""
        val name = "$firstName $lastName".trim()

        val result = buildJsonObject {
            put("id", body["id"]?.toString() ?: "")
            put("name", name)
            put("email", body["email"]?.toString()?.lowercase() ?: "")
        }

        serializeResponse(Response(200, result.toString()))
    }
    FunctionFly.run(transformer)
}
```

## Environment Variables

Access environment variables using `FunctionFly.getenv`:

```kotlin
val apiKey = FunctionFly.getenv("API_KEY")
val dbUrl = FunctionFly.getenv("DATABASE_URL")
val debug = FunctionFly.getenv("DEBUG") == "true"
```

## Security

The Kotlin runtime implements multiple security layers:

### Blocked Packages

The following packages are blocked by default:

| Package | Reason |
|---------|--------|
| `java.lang.Process` | Process execution |
| `java.lang.Runtime` | JVM control |
| `java.lang.System` | System access (exit, etc.) |
| `java.io.File` | File system access |
| `java.nio.file` | Advanced file ops |
| `java.net.Socket` | Network sockets |
| `java.net.ServerSocket` | Server sockets |
| `java.lang.ClassLoader` | Class loading |
| `sun.misc` | Internal APIs |
| `jdk.internal` | Internal APIs |

### Blocked Hosts

Cloud metadata endpoints are blocked:

- `169.254.169.254` (AWS)
- `metadata.google.internal` (GCP)
- `metadata.azure.com` (Azure)
- `100.100.100.200` (Alibaba Cloud)

### Environment Variable Restrictions

Sensitive variables are blocked:

- Variables containing `PASSWORD`, `SECRET`, `TOKEN`, `API_KEY`
- Variables with prefix `AWS_`, `GCP_`, `AZURE_`
- Variables like `LD_LIBRARY_PATH`, `DYLD_`, `LD_PRELOAD`

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 1024 MB |
| CPU | 10s | 60s |
| Output Size | 1 MB | 10 MB |
| Threads | 4 | 16 |

Configure in `functionfly.jsonc`:

```jsonc
{
  "runtime": "kotlin-jvm",
  "limits": {
    "timeout": 60,
    "memory": 512,
    "maxThreads": 8
  },
  "security": {
    "allowDiskIo": false,
    "allowNet": true
  }
}
```

## API Endpoints

The runtime exposes the following HTTP endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/ready` | GET | Readiness check |
| `/metrics` | GET | Metrics (JSON) |
| `/metrics/prom` | GET | Prometheus format |
| `/execute` | POST | Execute Kotlin code |
| `/validate` | POST | Validate code |

### Execute Request

```json
{
  "id": "uuid",
  "code": "fun main() { ... }",
  "entry": "main",
  "input": {"key": "value"},
  "timeout": 30000
}
```

### Execute Response

```json
{
  "id": "uuid",
  "success": true,
  "output": {"result": "data"},
  "execution_time_ms": 45,
  "memory_used_mb": 128,
  "terminated": false
}
```

## Cold Start

Kotlin/JVM functions have moderate cold starts:

- First invocation: ~100-500ms (JVM startup)
- Subsequent invocations: ~5-50ms
- Warm instances stay ready for faster responses

## Metrics

The runtime exports Prometheus metrics:

- `kotlin_runtime_total_executions` - Total executions
- `kotlin_runtime_successful_executions` - Successful executions
- `kotlin_runtime_failed_executions` - Failed executions
- `kotlin_runtime_timed_out_executions` - Timeouts
- `kotlin_runtime_security_violations_total` - Blocked security violations
- `kotlin_runtime_current_memory_mb` - Current memory usage
- `kotlin_runtime_peak_memory_mb` - Peak memory usage
- `kotlin_runtime_currently_executing` - Active executions
- `kotlin_runtime_uptime_seconds` - Runtime uptime

## Best Practices

1. **Keep functions small** - Reduce cold start time
2. **Avoid heavy dependencies** - Minimize JAR size
3. **Use `kotlinx.serialization`** - For JSON handling
4. **Handle exceptions gracefully** - Return appropriate status codes
5. **Set appropriate timeouts** - Match your function's needs
6. **Monitor memory usage** - Stay within limits
7. **Use environment variables** - For configuration, not secrets
8. **Return valid JSON** - Always serialize responses properly