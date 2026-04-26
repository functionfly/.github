/*
 * FunctionFly C SDK — Example Function
 *
 * Compile: emcc main.c -o function.wasm -s WASM=1 -s STANDALONE_WASM=1
 *   or:    clang --target=wasm32-wasi main.c -o function.wasm
 */
#include "functionfly.h"
#include <stdio.h>
#include <string.h>

/* Simple handler that returns a JSON greeting */
FF_EXPORT const char* execute(const char* input, int32_t input_len) {
    ff_log("Hello from C function!");

    /* Build response (static buffer for simplicity) */
    static char response[512];
    snprintf(response, sizeof(response),
        "{\"message\": \"Hello from C!\", \"input_length\": %d}",
        input_len);

    return response;
}
