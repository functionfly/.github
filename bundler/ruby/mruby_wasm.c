/*
 * mruby_wasm.c — WASM interface wrapper for the mruby interpreter.
 *
 * This file is compiled and linked with libmruby.a to produce mruby.wasm.
 * It exports a minimal set of functions for the FunctionFly runtime:
 *
 *   mruby_init()                  — create an mruby state
 *   mruby_exec(code_ptr, code_len) — load & run Ruby source from WASM memory
 *   mruby_exec_string(code_ptr)   — load & run null-terminated Ruby source
 *   mruby_result_ptr()            — pointer to the last result string
 *   mruby_result_len()            — length of the last result string
 *   mruby_error_ptr()             — pointer to the last error string (or 0)
 *   mruby_error_len()             — length of the last error string
 *   mruby_cleanup()               — free the mruby state
 *   malloc / free                 — host memory management
 *
 * The host (wasmtime) writes Ruby source into WASM linear memory via malloc,
 * calls mruby_exec, then reads the result/error from the exported pointers.
 */

#include <mruby.h>
#include <mruby/compile.h>
#include <mruby/string.h>
#include <mruby/value.h>
#include <mruby/error.h>
#include <string.h>
#include <stdlib.h>

/* ── state ───────────────────────────────────────────────────────────────── */

static mrb_state *g_mrb = NULL;

/* result / error buffers — fixed-size, exported via accessor functions */
#define RESULT_BUF_SIZE 65536
static char g_result_buf[RESULT_BUF_SIZE];
static int  g_result_len = 0;

static char g_error_buf[RESULT_BUF_SIZE];
static int  g_error_len = 0;

/* ── exports ─────────────────────────────────────────────────────────────── */

/*
 * mruby_init — create a new mrb_state.
 * Returns 0 on success, -1 if already initialised, -1 on alloc failure.
 */
int mruby_init(void) {
    if (g_mrb != NULL) {
        return -1;  /* already initialised */
    }
    g_mrb = mrb_open();
    if (g_mrb == NULL) {
        return -1;
    }
    g_result_len = 0;
    g_error_len  = 0;
    g_result_buf[0] = '\0';
    g_error_buf[0]  = '\0';
    return 0;
}

/*
 * mruby_exec — execute Ruby source code.
 *   code_ptr  — pointer to UTF-8 Ruby source in WASM linear memory
 *   code_len  — byte length of the source (may contain NULs)
 * Returns 0 on success, non-zero on error.
 */
int mruby_exec(int code_ptr, int code_len) {
    if (g_mrb == NULL) {
        return -1;
    }

    const char *src = (const char *)(uintptr_t)code_ptr;
    if (src == NULL || code_len <= 0) {
        return -1;
    }

    /* mruby_load_nstring expects a length-bounded string */
    mrb_value result = mrb_load_nstring(g_mrb, src, code_len);

    g_result_len = 0;
    g_error_len  = 0;

    if (g_mrb->exc) {
        /* Exception occurred — capture message */
        mrb_value exc = mrb_obj_value(g_mrb->exc);
        mrb_value msg = mrb_funcall(g_mrb, exc, "message", 0);

        if (mrb_string_p(msg)) {
            const char *s = mrb_str_to_cstr(g_mrb, msg);
            int len = (int)strlen(s);
            if (len >= RESULT_BUF_SIZE) len = RESULT_BUF_SIZE - 1;
            memcpy(g_error_buf, s, len);
            g_error_buf[len] = '\0';
            g_error_len = len;
        } else {
            const char *fallback = "unknown mruby error";
            int len = (int)strlen(fallback);
            memcpy(g_error_buf, fallback, len);
            g_error_buf[len] = '\0';
            g_error_len = len;
        }

        g_mrb->exc = NULL;
        return 1;
    }

    /* Success — convert result to string */
    mrb_value str = mrb_funcall(g_mrb, result, "to_s", 0);
    if (mrb_string_p(str)) {
        const char *s = mrb_str_to_cstr(g_mrb, str);
        int len = (int)strlen(s);
        if (len >= RESULT_BUF_SIZE) len = RESULT_BUF_SIZE - 1;
        memcpy(g_result_buf, s, len);
        g_result_buf[len] = '\0';
        g_result_len = len;
    } else {
        g_result_buf[0] = '\0';
        g_result_len = 0;
    }

    return 0;
}

/*
 * mruby_exec_string — execute a null-terminated Ruby source string.
 * Convenience wrapper around mruby_exec.
 */
int mruby_exec_string(int code_ptr) {
    const char *src = (const char *)(uintptr_t)code_ptr;
    if (src == NULL) return -1;
    return mruby_exec(code_ptr, (int)strlen(src));
}

/* ── result accessors ────────────────────────────────────────────────────── */

int mruby_result_ptr(void) { return (int)(uintptr_t)g_result_buf; }
int mruby_result_len(void) { return g_result_len; }
int mruby_error_ptr(void)  { return (int)(uintptr_t)g_error_buf; }
int mruby_error_len(void)  { return g_error_len; }

/*
 * mruby_cleanup — close the mruby state and free resources.
 */
void mruby_cleanup(void) {
    if (g_mrb != NULL) {
        mrb_close(g_mrb);
        g_mrb = NULL;
    }
    g_result_len = 0;
    g_error_len  = 0;
    g_result_buf[0] = '\0';
    g_error_buf[0]  = '\0';
}
