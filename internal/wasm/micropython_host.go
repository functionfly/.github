//go:build cgo

package wasm

import (
	"github.com/bytecodealliance/wasmtime-go/v19"
)

// DefineMicropythonHostFunctions defines the env.* functions that micropython.wasm needs
// These are JavaScript interop stubs that aren't needed for serverless execution
func DefineMicropythonHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store) error {
	// env.invoke_* stubs
	if err := linker.DefineFunc(store, "env", "invoke_ii", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iiii", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_v", func(caller *wasmtime.Caller, a int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_viii", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iiiii", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iii", func(caller *wasmtime.Caller, a, b, c int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_vi", func(caller *wasmtime.Caller, a int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_vii", func(caller *wasmtime.Caller, a, b int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_i", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}

	// env.mp_js_* stubs
	if err := linker.DefineFunc(store, "env", "mp_js_hook", func(caller *wasmtime.Caller, a int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_random_u32", func(caller *wasmtime.Caller) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_ticks_ms", func(caller *wasmtime.Caller) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_time_ms", func(caller *wasmtime.Caller) float64 { return 0 }); err != nil {
		return err
	}

	// env.emscripten_* stubs
	if err := linker.DefineFunc(store, "env", "emscripten_scan_registers", func(caller *wasmtime.Caller, a int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "emscripten_resize_heap", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "_emscripten_throw_longjmp", func(caller *wasmtime.Caller) {}); err != nil {
		return err
	}

	// env.proxy_* stubs
	if err := linker.DefineFunc(store, "env", "proxy_convert_mp_to_js_then_js_to_mp_obj_jsside", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "proxy_convert_mp_to_js_then_js_to_js_then_js_to_mp_obj_jsside", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_proxy_js_ref_info", func(caller *wasmtime.Caller, a int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_iter", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "proxy_js_free_obj", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_reflect_construct", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_iter_next", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_check_existing", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_error_info", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_resolve", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "create_promise", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_continue", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_reject", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}

	// env.call* stubs
	if err := linker.DefineFunc(store, "env", "call0_kwarg", func(caller *wasmtime.Caller, a, b, c, d, e, f, g, h int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "calln_kwarg", func(caller *wasmtime.Caller, a, b, c, d, e, f, g, h int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call1", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call2", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "calln", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call0", func(caller *wasmtime.Caller, a, b, c int32) {}); err != nil {
		return err
	}

	// env.lookup_attr, env.store_attr stubs
	if err := linker.DefineFunc(store, "env", "lookup_attr", func(caller *wasmtime.Caller, a, b, c, d int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "store_attr", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_subscr_load", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_subscr_store", func(caller *wasmtime.Caller, a, b, c, d int32) {}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "has_attr", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}

	// env.__syscall_* stubs
	if err := linker.DefineFunc(store, "env", "__syscall_chdir", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_getcwd", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_mkdirat", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_openat", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_poll", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_getdents64", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_renameat", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_rmdir", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_fstat64", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_stat64", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_newfstatat", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_lstat64", func(caller *wasmtime.Caller, a int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_statfs64", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "__syscall_unlinkat", func(caller *wasmtime.Caller, a, b int32) int32 { return 0 }); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "_abort_js", func(caller *wasmtime.Caller) {}); err != nil {
		return err
	}

return nil
}
