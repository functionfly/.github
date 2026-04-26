---
title: Kotlin SDK
description: Kotlin SDK for building FunctionFly functions
---

# Kotlin SDK

The FunctionFly Kotlin SDK provides a functional API for building FunctionFly functions in Kotlin, compiled to WebAssembly via Kotlin/WASM.

## Installation

Add the FunctionFly SDK to your `build.gradle.kts`:

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
    sourceSets {
        val wasmWasiMain by getting {
            dependencies {
                implementation("com.functionfly:sdk-kotlin:1.0.0")
            }
        }
    }
}
```

## Quick Start

```kotlin
import functionfly.*

fun main() {
    val myFunction = Function { input ->
        """{"message": "Hello from Kotlin!"}"""
    }
    FunctionFly.run(myFunction)
}
```

## Function Interface

All functions implement the `Function` functional interface:

```kotlin
fun interface Function {
    fun handle(input: String): String
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

```kotlin
import functionfly.*
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
                }
                buildJsonObject {
                    put("status", 200)
                    put("body", buildJsonObject { put("users", users) })
                }.toString()
            }
            "POST" -> {
                val body = request["body"]?.jsonObject
                buildJsonObject {
                    put("status", 201)
                    put("body", buildJsonObject {
                        put("id", "3")
                        put("name", body?.get("name")?.jsonPrimitive?.content ?: "")
                        put("email", body?.get("email")?.jsonPrimitive?.content ?: "")
                    })
                }.toString()
            }
            else -> """{"status": 405, "body": {"error": "Method not allowed"}}"""
        }
    }

    FunctionFly.run(api)
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

### Using Environment Variables

```kotlin
import functionfly.*

fun main() {
    val fn = Function { input ->
        val apiKey = FunctionFly.getEnv("API_KEY")
        val debug = FunctionFly.getEnv("DEBUG") == "true"

        FunctionFly.log("Processing with debug=$debug")

        """{"status": 200, "api_key_set": ${apiKey.isNotEmpty()}, "debug": $debug}"""
    }

    FunctionFly.run(fn)
}
```

## JSON Handling

The SDK works with `kotlinx.serialization` for JSON:

```kotlin
import kotlinx.serialization.json.*

// Parse
val obj = Json.parseToJsonElement(jsonString).jsonObject

// Build
val response = buildJsonObject {
    put("status", 200)
    put("body", buildJsonObject {
        put("key", "value")
        put("count", 42)
    })
}.toString()
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

## Building for WASM

```bash
# Build WASM binary
./gradlew wasmWasiBinaries

# Output location
# build/bin/wasmWasi/releaseExecutable/my-function.wasm
```

## API Reference

### Interfaces

| Interface | Method | Signature |
|-----------|--------|-----------|
| `Function` | `handle` | `fun handle(input: String): String` |

### Module Functions

| Function | Parameters | Returns | Description |
|----------|-----------|---------|-------------|
| `FunctionFly.run` | `function: Function` | `void` | Register and execute |
| `FunctionFly.log` | `message: String` | `void` | Log message |
| `FunctionFly.getEnv` | `key: String` | `String` | Get env variable |
