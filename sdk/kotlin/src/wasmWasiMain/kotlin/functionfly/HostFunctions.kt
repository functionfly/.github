package functionfly

/**
 * Host function declarations for the FunctionFly WASM runtime.
 * These are provided by the runtime via WASI imports.
 */

/** Log a message via the FunctionFly logging host function. */
external fun functionflyLog(msg: String)

/** Fetch a URL via the FunctionFly HTTP proxy. */
external fun functionflyFetch(url: String): String?

/** Get a value from the key-value store. */
external fun functionflyKvGet(key: String): String?

/** Set a value in the key-value store. */
external fun functionflyKvSet(key: String, value: String): Boolean

/** Run AI inference. */
external fun functionflyAi(prompt: String): String?
