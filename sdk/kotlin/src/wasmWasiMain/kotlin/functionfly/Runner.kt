package functionfly

/**
 * Helper to run a FunctionFly function as a Kotlin/WASM executable.
 */
fun runFunction(function: Function) {
    val ctx = Context.create()
    val input = readStdin()
    val output = try {
        function.handle(input)
    } catch (e: Exception) {
        ctx.error("RUNTIME_ERROR", e.message ?: "Unknown error")
    }
    print(output)
}

private fun readStdin(): String {
    // In WASI, stdin is available via standard input
    return generateSequence { readlnOrNull() }.joinToString("\n")
}
