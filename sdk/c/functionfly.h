/*
 * FunctionFly C SDK — Header
 *
 * Provides types and host function declarations for building FunctionFly
 * functions in C. Compile with Emscripten or WASI-SDK targeting wasm32.
 *
 * Usage:
 *   #include "functionfly.h"
 *
 *   FF_EXPORT char* handler(const char* input) {
 *       return ff_strdup("{\"message\": \"Hello from C!\"}");
 *   }
 */
#ifndef FUNCTIONFLY_H
#define FUNCTIONFLY_H

#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <stdlib.h>

/* ── Export macros ──────────────────────────────────────────────── */

#ifdef __EMSCRIPTEN__
#include <emscripten.h>
#define FF_EXPORT EMSCRIPTEN_KEEPALIVE
#else
#define FF_EXPORT __attribute__((visibility("default")))
#endif

/* ── Standard FunctionFly entry points ─────────────────────────── */

#ifdef __cplusplus
extern "C" {
#endif

/* Called once at cold start */
FF_EXPORT void init(void);

/* Main execution: receives input string, returns output string */
FF_EXPORT const char* execute(const char* input, int32_t input_len);

/* Memory management */
FF_EXPORT void* alloc(int32_t size);
FF_EXPORT void  dealloc(void* ptr);

/* Returns JSON metadata about this function */
FF_EXPORT const char* metadata(void);

#ifdef __cplusplus
}
#endif

/* ── Host function imports (provided by FunctionFly runtime) ───── */

#ifdef __cplusplus
extern "C" {
#endif

/* Logging */
extern void functionfly_log(const char* msg, int32_t len);

/* Environment variables */
extern int32_t functionfly_get_env(const char* key, int32_t key_len, char* buf, int32_t buf_len);

/* HTTP fetch (requires fetch:read/write capability) */
extern int32_t functionfly_fetch(const char* url, int32_t url_len, char* buf, int32_t buf_len);

/* Key-value store */
extern int32_t functionfly_kv_get(const char* key, int32_t key_len, char* buf, int32_t buf_len);
extern int32_t functionfly_kv_set(const char* key, int32_t key_len, const char* val, int32_t val_len);

/* AI inference (requires ai capability) */
extern int32_t functionfly_ai(const char* prompt, int32_t prompt_len, char* buf, int32_t buf_len);

/* Crypto (requires crypto capability) */
extern int32_t functionfly_crypto_hash(const char* algo, int32_t algo_len, const char* data, int32_t data_len, char* buf, int32_t buf_len);

#ifdef __cplusplus
}
#endif

/* ── Utility functions ─────────────────────────────────────────── */

static inline char* ff_strdup(const char* s) {
    if (!s) return NULL;
    size_t len = strlen(s);
    char* dup = (char*)malloc(len + 1);
    if (dup) {
        memcpy(dup, s, len + 1);
    }
    return dup;
}

static inline int32_t ff_string_len(const char* s) {
    return s ? (int32_t)strlen(s) : 0;
}

/* Helper to call functionfly_log with a C string */
static inline void ff_log(const char* msg) {
    if (msg) {
        functionfly_log(msg, (int32_t)strlen(msg));
    }
}

#endif /* FUNCTIONFLY_H */
