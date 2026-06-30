/*
 * FunctionFly example: Hello from C
 * Build: emcc hello.c -o hello.wasm -s WASM=1 -s STANDALONE_WASM=1 --no-entry
 *    or: clang --target=wasm32-wasi hello.c -o hello.wasm -nostartfiles
 * Manifest: {"name":"hello-c","version":"1.0.0","runtime":"c","entry":"hello.c"}
 */
#include <stdio.h>
#include <string.h>

__attribute__((visibility("default")))
void init(void) {
    /* no-op */
}

__attribute__((visibility("default")))
const char* execute(const char* input, int input_len) {
    static char response[512];
    snprintf(response, sizeof(response),
        "{\"message\": \"Hello from C!\", \"input_length\": %d, \"ok\": true}",
        input_len);
    return response;
}
