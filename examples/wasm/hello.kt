// FunctionFly example: Hello from Kotlin
// Build: gradle wasmWasiNodeProductionRun
// Manifest: {"name":"hello-kotlin","version":"1.0.0","runtime":"kotlin","entry":"Main.kt"}

fun main() {
    val input = generateSequence { readlnOrNull() }.joinToString("\n")
    val response = """{"message": "Hello from Kotlin!", "input_length": ${input.length}, "ok": true}"""
    print(response)
}
