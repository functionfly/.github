/*
 * FunctionFly C SDK — Implementation
 *
 * Provides default implementations for the standard FunctionFly entry points.
 * Override these in your function source file.
 */
#include "functionfly.h"

/* Default init — override in your function */
__attribute__((weak))
FF_EXPORT void init(void) {
    /* no-op by default */
}

/* Default metadata — override to provide custom metadata */
__attribute__((weak))
FF_EXPORT const char* metadata(void) {
    return "{\"runtime\":\"c\",\"version\":\"1.0.0\"}";
}

/* Default alloc/dealloc using stdlib */
FF_EXPORT void* alloc(int32_t size) {
    return malloc((size_t)size);
}

FF_EXPORT void dealloc(void* ptr) {
    free(ptr);
}
