---
title: Kotlin via WASM Runtime
description: Kotlin WebAssembly runtime environment for FunctionFly functions
---

# Kotlin via WASM Runtime

FunctionFly's Kotlin runtime compiles your Kotlin code to WebAssembly using Kotlin/WASM for type-safe, sandboxed execution.

## Overview

Kotlin functions are compiled to WebAssembly using JetBrains' Kotlin/WASM compiler. This provides:

- **Type safety** - Full Kotlin type system with null safety
- **Fast cold starts** - WASM modules initialize quickly
- **Sandboxed** - Memory-safe, isolated execution
- **Coroutines support** - Async operations via Kotlin coroutines
- **Familiar ecosystem** - Standard Kotlin tooling and IDE support

## Supported Toolchains

| Toolchain | Status | Notes |
|-----------|--------|-------|
| Kotlin/WASM (wasmWasi) | Supported | Recommended; direct compilation |
| Kotlin → JS → Javy | Supported | Fallback; via Javy WASM runtime |

## Project Structure

```
my-function/
├── build.gradle.kts        # Gradle build config
├── settings.gradle.kts     # Project settings
├── src/
│   └── wasmWasiMain/
│       └── kotlin/
│           └── Main.kt     # Entry point
└── functionfly.jsonc       # Function config
```

## Build Configuration

```kotlin
// build.gradle.kts
plugins {
    kotlin("multiplatform") version "2.1.0"
}

kotlin {
    wasmWasi {
        binaries {
            executable()
        }
    }
}
```

## Function Structure

A Kotlin function implements the `FunctionFly.Function` interface:

```kotlin
package functionfly

fun interface Function {
    fun handle(input: String): String
}
```

Basic example:

```kotlin
import functionfly.*

fun main() {
    val myFunction = Function { input ->
        """{"message": "Hello from Kotlin!"}"""
    }
    FunctionFly.run(myFunction)
}
```

## Example Functions

### HTTP API Handler

```kotlin
import functionfly.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*

fun main() {
    val api = Function { input ->
        val request = Json.parseToJsonElement(input).jsonObject
        val method = request["method"]?.jsonPrimitive?.content ?: "GET"

        when (method) {
            "GET" -> {
                val users = buildJsonArray {
                    addJsonObject {
                        put("id", "1")
                        put("name", "Alice")
                        put("email", "alice@example.com")
                    }
                    addJsonObject {
                        put("id", "2")
                        put("name", "Bob")
                        put("email", "bob@example.com")
                    }
                }
                buildJsonObject {
                    put("status", 200)
                    put("body", buildJsonObject { put("users", users) })
                }.toString()
            }
            "POST" -> {
                val body = request["body"]?.jsonObject
                val name = body?.get("name")?.jsonPrimitive?.content ?: ""
                val email = body?.get("email")?.jsonPrimitive?.content ?: ""
                buildJsonObject {
                    put("status", 201)
                    put("body", buildJsonObject {
                        put("id", "3")
                        put("name", name)
                        put("email", email)
                    })
                }.toString()
            }
            else -> """{"status": 405, "body": {"error": "Method not allowed"}}"""
        }
    }

    FunctionFly.run(api)
}
```

### Webhook Processor

```kotlin
import functionfly.*

fun main() {
    val webhook = Function { input ->
        val request = parseJson(input)
        val secret = FunctionFly.getEnv("WEBHOOK_SECRET")
        val signature = request["headers"]?.get("x-signature") ?: ""

        if (!verifySignature(request["body"].toString(), signature, secret)) {
            return@Function """{"status": 401, "body": {"error": "Invalid signature"}}"""
        }

        FunctionFly.log("Webhook received")
        """{"status": 200, "body": {"received": true}}"""
    }

    FunctionFly.run(webhook)
}

fun verifySignature(body: String, signature: String, secret: String): Boolean {
    if (!signature.startsWith("sha256=")) return false
    // In production, compute HMAC-SHA256 and compare
    return true
}
```

### Data Transformation

```kotlin
import functionfly.*
import kotlinx.serialization.json.*

fun main() {
    val transformer = Function { input ->
        val request = Json.parseToJsonElement(input).jsonObject
        val body = request["body"]?.jsonObject ?: return@Function """{"status": 400}"""

        val firstName = body["first_name"]?.jsonPrimitive?.content ?: ""
        val lastName = body["last_name"]?.jsonPrimitive?.content ?: ""
        val name = "$firstName $lastName".trim()

        buildJsonObject {
            put("status", 200)
            put("body", buildJsonObject {
                put("id", body["id"]?.jsonPrimitive?.content ?: "")
                put("name", name)
                put("email", body["email"]?.jsonPrimitive?.content?.lowercase() ?: "")
                put("timestamp", body["created_at"]?.jsonPrimitive?.content ?: "")
            })
        }.toString()
    }

    FunctionFly.run(transformer)
}
```

## Environment Variables

Access environment variables using `FunctionFly.getEnv`:

```kotlin
val apiKey = FunctionFly.getEnv("API_KEY")
val dbUrl = FunctionFly.getEnv("DATABASE_URL")
val debug = FunctionFly.getEnv("DEBUG") == "true"
```

## Error Handling

```kotlin
val safeFunction = Function { input ->
    try {
        val request = Json.parseToJsonElement(input).jsonObject
        val result = processRequest(request)
        """{"status": 200, "body": $result}"""
    } catch (e: IllegalArgumentException) {
        """{"status": 400, "body": {"error": "${e.message}"}}"""
    } catch (e: Exception) {
        FunctionFly.log("Error: ${e.message}")
        """{"status": 500, "body": {"error": "${e.message}"}}"""
    }
}
```

## functionfly.jsonc Configuration

```jsonc
{
  "name": "my-kotlin-function",
  "runtime": "kotlin-wasm",
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

Kotlin/WASM functions may use more memory than C/Rust due to the Kotlin runtime.

## Cold Start

Kotlin/WASM functions have moderate cold starts:
- First invocation after deployment: ~30-100ms
- Subsequent invocations: ~3-10ms

## Best Practices

1. **Use `kotlinx.serialization`** - For JSON parsing and generation
2. **Keep the runtime small** - Avoid unnecessary dependencies
3. **Handle nulls safely** - Use Kotlin's null safety features
4. **Use `buildJsonObject`** - For constructing JSON responses
5. **Log errors** - Use `FunctionFly.log` for debugging
6. **Return valid JSON** - Always return a JSON string from the function
7. **Optimize binary size** - Enable R8/ProGuard shrinking in release builds
