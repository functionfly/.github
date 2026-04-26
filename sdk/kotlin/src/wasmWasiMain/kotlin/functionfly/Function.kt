package functionfly

/**
 * Interface that all FunctionFly Kotlin functions must implement.
 */
fun interface Function {
    /**
     * Process the input and return the output.
     * @param input The function input as a JSON string
     * @return The function output as a JSON string
     */
    fun handle(input: String): String
}
