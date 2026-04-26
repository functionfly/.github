package functionfly

/**
 * Example FunctionFly function in Kotlin.
 * Build: gradle wasmWasiNodeProductionRun
 */
fun main() {
    runFunction(Function { input ->
        """{"message": "Hello from Kotlin!", "ok": true}"""
    })
}
