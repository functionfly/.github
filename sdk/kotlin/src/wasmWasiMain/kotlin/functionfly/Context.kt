package functionfly

/**
 * Execution context for a FunctionFly function.
 * Provides access to runtime environment and host functions.
 */
class Context(
    val functionName: String = "",
    val version: String = ""
) {
    companion object {
        fun create(): Context {
            return Context(
                functionName = getEnv("FUNCTIONFLY_FUNCTION_NAME") ?: "",
                version = getEnv("FUNCTIONFLY_VERSION") ?: ""
            )
        }
    }

    /** Log a message to FunctionFly logging. */
    fun log(message: String) {
        println("[functionfly] $message")
    }

    /** Return an error JSON response. */
    fun error(code: String, message: String): String {
        return """{"error": {"code": "$code", "message": "$message"}}"""
    }

    /** Return a success JSON response. */
    fun json(data: String): String {
        return """{"ok": true, "data": $data}"""
    }
}

/** Get environment variable (external WASI function). */
external fun getEnv(key: String): String?
