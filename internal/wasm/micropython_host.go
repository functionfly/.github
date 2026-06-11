//go:build cgo

package wasm

import (
	"github.com/sirupsen/logrus"
	"github.com/bytecodealliance/wasmtime-go/v19"
)

// stubSecurityEnabled controls whether to log stub function calls (audit purposes)
const stubSecurityEnabled = true

// DefineMicropythonHostFunctions defines the env.* functions that micropython.wasm needs
// These are JavaScript interop stubs that aren't needed for serverless execution
// For streaming support, use DefineMicropythonHostFunctionsWithState
func DefineMicropythonHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store) error {
	return DefineMicropythonHostFunctionsWithState(linker, store, nil, nil)
}

// DefineMicropythonHostFunctionsWithState defines the env.* functions with streaming state support
// If streamingState is nil, streaming functions will be stub implementations
// If getMemory is nil, a default memory accessor will be used
func DefineMicropythonHostFunctionsWithState(linker *wasmtime.Linker, store *wasmtime.Store, streamingState *StreamingState, getMemory func() []byte) error {
	// env.invoke_* stubs - with security hardening for pointer validation
	if err := linker.DefineFunc(store, "env", "invoke_ii", func(caller *wasmtime.Caller, a, b int32) int32 {
		if stubSecurityEnabled && (a < 0 || b < 0) {
			logrus.WithFields(logrus.Fields{
				"function": "invoke_ii",
				"a":        a,
				"b":        b,
			}).Warn("[Security] invoke_ii: negative pointer(s)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iiii", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "invoke_iiii"}).Warn("[Security] invoke_iiii: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_v", func(caller *wasmtime.Caller, a int32) {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "invoke_v", "a": a}).Warn("[Security] invoke_v: negative pointer")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_viii", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "invoke_viii"}).Warn("[Security] invoke_viii: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iiiii", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0 || e < 0) {
			logrus.WithFields(logrus.Fields{"function": "invoke_iiiii"}).Warn("[Security] invoke_iiiii: negative pointer(s)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_iii", func(caller *wasmtime.Caller, a, b, c int32) int32 {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0) {
			logrus.WithFields(logrus.Fields{"function": "invoke_iii"}).Warn("[Security] invoke_iii: negative pointer(s)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_vi", func(caller *wasmtime.Caller, a int32) {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "invoke_vi", "a": a}).Warn("[Security] invoke_vi: negative pointer")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_vii", func(caller *wasmtime.Caller, a, b int32) {
		if stubSecurityEnabled && (a < 0 || b < 0) {
			logrus.WithFields(logrus.Fields{"function": "invoke_vii"}).Warn("[Security] invoke_vii: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "invoke_i", func(caller *wasmtime.Caller, a int32) int32 {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "invoke_i", "a": a}).Warn("[Security] invoke_i: negative pointer")
		}
		return 0
	}); err != nil {
		return err
	}

	// env.mp_js_* stubs - with security hardening
	if err := linker.DefineFunc(store, "env", "mp_js_hook", func(caller *wasmtime.Caller, a int32) {
		if stubSecurityEnabled {
			mem := caller.GetExport("memory")
			if mem != nil {
				memory := mem.Memory()
				if memory != nil {
					memData := memory.UnsafeData(store)
					if a < 0 || int(a) >= len(memData) {
						logrus.WithFields(logrus.Fields{"function": "mp_js_hook", "a": a}).Warn("[Security] mp_js_hook: invalid pointer")
						return
					}
				}
			}
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_random_u32", func(caller *wasmtime.Caller) int32 {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "mp_js_random_u32"}).Warn("[Security] mp_js_random_u32 called (returns 0 - no entropy source)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_ticks_ms", func(caller *wasmtime.Caller) int32 {
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "mp_js_time_ms", func(caller *wasmtime.Caller) float64 {
		return 0
	}); err != nil {
		return err
	}

	// env.emscripten_* stubs - with security hardening
	if err := linker.DefineFunc(store, "env", "emscripten_scan_registers", func(caller *wasmtime.Caller, a int32) {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "emscripten_scan_registers", "a": a}).Warn("[Security] emscripten_scan_registers: negative pointer")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "emscripten_resize_heap", func(caller *wasmtime.Caller, a int32) int32 {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "emscripten_resize_heap", "size": a}).Warn("[Security] emscripten_resize_heap called (rejected - no dynamic memory)")
		}
		return -1
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "_emscripten_throw_longjmp", func(caller *wasmtime.Caller) {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "_emscripten_throw_longjmp"}).Info("[Security] _emscripten_throw_longjmp called (no-op in serverless)")
		}
	}); err != nil {
		return err
	}

	// env.proxy_* stubs - with security hardening for pointer validation
	if err := linker.DefineFunc(store, "env", "proxy_convert_mp_to_js_then_js_to_mp_obj_jsside", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0 || e < 0 || f < 0) {
			logrus.WithFields(logrus.Fields{"function": "proxy_convert_mp_to_js"}).Warn("[Security] proxy_convert_mp_to_js: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "proxy_convert_mp_to_js_then_js_to_js_then_js_to_mp_obj_jsside", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0 || e < 0 || f < 0) {
			logrus.WithFields(logrus.Fields{"function": "proxy_convert_mp_to_js_chain"}).Warn("[Security] proxy_convert_mp_to_js_chain: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_proxy_js_ref_info", func(caller *wasmtime.Caller, a int32) {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "js_get_proxy_js_ref_info", "a": a}).Warn("[Security] js_get_proxy_js_ref_info: negative pointer")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_iter", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_get_iter"}).Warn("[Security] js_get_iter: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "proxy_js_free_obj", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "proxy_js_free_obj"}).Warn("[Security] proxy_js_free_obj: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_reflect_construct", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0 || e < 0 || f < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_reflect_construct"}).Warn("[Security] js_reflect_construct: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_iter_next", func(caller *wasmtime.Caller, a int32) int32 {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "js_iter_next", "a": a}).Warn("[Security] js_iter_next: negative pointer")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_check_existing", func(caller *wasmtime.Caller, a int32) int32 {
		if stubSecurityEnabled && a < 0 {
			logrus.WithFields(logrus.Fields{"function": "js_check_existing", "a": a}).Warn("[Security] js_check_existing: negative pointer")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_get_error_info", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_get_error_info"}).Warn("[Security] js_get_error_info: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_resolve", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_then_resolve"}).Warn("[Security] js_then_resolve: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "create_promise", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "create_promise"}).Warn("[Security] create_promise: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_continue", func(caller *wasmtime.Caller, a, b, c, d, e, f int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0 || e < 0 || f < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_then_continue"}).Warn("[Security] js_then_continue: negative pointer(s)")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "js_then_reject", func(caller *wasmtime.Caller, a, b, c, d int32) {
		if stubSecurityEnabled && (a < 0 || b < 0 || c < 0 || d < 0) {
			logrus.WithFields(logrus.Fields{"function": "js_then_reject"}).Warn("[Security] js_then_reject: negative pointer(s)")
		}
	}); err != nil {
		return err
	}

	// env.call* stubs - with security hardening
	if err := linker.DefineFunc(store, "env", "call0_kwarg", func(caller *wasmtime.Caller, a, b, c, d, e, f, g, h int32) {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "call0_kwarg"}).Debug("[Security] call0_kwarg: no-op in serverless context")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "calln_kwarg", func(caller *wasmtime.Caller, a, b, c, d, e, f, g, h int32) {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "calln_kwarg"}).Debug("[Security] calln_kwarg: no-op in serverless context")
		}
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call1", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "call1"}).Debug("[Security] call1: returns 0 (not implemented)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call2", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "call2"}).Debug("[Security] call2: returns 0 (not implemented)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "calln", func(caller *wasmtime.Caller, a, b, c, d, e int32) int32 {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "calln"}).Debug("[Security] calln: returns 0 (not implemented)")
		}
		return 0
	}); err != nil {
		return err
	}
	if err := linker.DefineFunc(store, "env", "call0", func(caller *wasmtime.Caller, a, b, c int32) {
		if stubSecurityEnabled {
			logrus.WithFields(logrus.Fields{"function": "call0"}).Debug("[Security] call0: no-op in serverless context")
		}
	}); err != nil {
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

	// Streaming support functions for chunked input/output
	// These allow the WASM module to receive input chunks and emit output chunks

	// streaming_init - Initialize streaming state
	if err := linker.DefineFunc(store, "env", "streaming_init", func(caller *wasmtime.Caller) int32 {
		if streamingState != nil {
			streamingState.Init()
		}
		return 0 // Success
	}); err != nil {
		return err
	}

	// streaming_send_chunk - WASM acknowledges receipt of a chunk
	if err := linker.DefineFunc(store, "env", "streaming_send_chunk", func(caller *wasmtime.Caller, chunkID int32, ptr int32, len int32, isLast int32) int32 {
		if streamingState == nil {
			return 0 // Stub: no state
		}
		// The chunk data is at memory[ptr:ptr+len]
		// WASM is telling us it received this chunk
		// For now, we just acknowledge - actual data is already in memory
		logrus.WithFields(logrus.Fields{"chunkID": chunkID, "len": len, "isLast": isLast != 0}).Debug("[Streaming] WASM acknowledged chunk")
		return 0 // Success
	}); err != nil {
		return err
	}

	// streaming_get_output_chunk - Get pointer to output chunk metadata for given chunk ID
	// Returns pointer to chunk metadata (chunkID, ptr, len) or 0 if not available
	// Metadata format: [4 bytes chunkID, 4 bytes ptr, 4 bytes len, 4 bytes isLast]
	if err := linker.DefineFunc(store, "env", "streaming_get_output_chunk", func(caller *wasmtime.Caller, chunkID int32) int32 {
		if streamingState == nil {
			return 0 // Stub
		}

		// Get memory to read output chunks
		mem := caller.GetExport("memory")
		if mem == nil {
			return 0
		}
		memory := mem.Memory()
		if memory == nil {
			return 0
		}
		memData := memory.UnsafeData(store)

		// Get the output chunk
		chunk, exists := streamingState.GetOutputChunk(chunkID)
		if !exists || chunk == nil {
			return 0 // No chunk available
		}

		// Write metadata to a known location (chunk metadata table at offset 4096)
		metaOffset := 4096 + int(chunkID)*16
		if metaOffset+16 > len(memData) {
			return 0 // Out of bounds
		}

		// Write chunk metadata
		ptr := streamingState.GetOutputChunkPtr(chunkID)

		// chunkID
		memData[metaOffset] = byte(chunkID)
		memData[metaOffset+1] = byte(chunkID >> 8)
		memData[metaOffset+2] = byte(chunkID >> 16)
		memData[metaOffset+3] = byte(chunkID >> 24)

		// ptr
		memData[metaOffset+4] = byte(ptr)
		memData[metaOffset+5] = byte(ptr >> 8)
		memData[metaOffset+6] = byte(ptr >> 16)
		memData[metaOffset+7] = byte(ptr >> 24)

		// len
		chunkLen := int32(len(chunk))
		memData[metaOffset+8] = byte(chunkLen)
		memData[metaOffset+9] = byte(chunkLen >> 8)
		memData[metaOffset+10] = byte(chunkLen >> 16)
		memData[metaOffset+11] = byte(chunkLen >> 24)

		// isLast (0 = false, 1 = true)
		memData[metaOffset+12] = 0
		memData[metaOffset+13] = 0
		memData[metaOffset+14] = 0
		memData[metaOffset+15] = 0

		logrus.WithFields(logrus.Fields{"chunkID": chunkID, "ptr": ptr, "len": chunkLen}).Debug("[Streaming] Output chunk")
		return int32(metaOffset)
	}); err != nil {
		return err
	}

	// streaming_get_input_chunk - Get pointer to input chunk metadata for given chunk ID
	// Returns pointer to chunk metadata or 0 if not available
	if err := linker.DefineFunc(store, "env", "streaming_get_input_chunk", func(caller *wasmtime.Caller, chunkID int32) int32 {
		if streamingState == nil {
			return 0 // Stub
		}

		// Get memory
		mem := caller.GetExport("memory")
		if mem == nil {
			return 0
		}
		memory := mem.Memory()
		if memory == nil {
			return 0
		}
		memData := memory.UnsafeData(store)

		// Get the input chunk
		chunk, isLast := streamingState.GetInputChunk(chunkID)
		if chunk == nil {
			return 0 // No chunk available
		}

		// Write metadata
		metaOffset := 8192 + int(chunkID)*16
		if metaOffset+16 > len(memData) {
			return 0
		}

		ptr := streamingState.GetInputChunkPtr(chunkID)

		// chunkID
		memData[metaOffset] = byte(chunkID)
		memData[metaOffset+1] = byte(chunkID >> 8)
		memData[metaOffset+2] = byte(chunkID >> 16)
		memData[metaOffset+3] = byte(chunkID >> 24)

		// ptr
		memData[metaOffset+4] = byte(ptr)
		memData[metaOffset+5] = byte(ptr >> 8)
		memData[metaOffset+6] = byte(ptr >> 16)
		memData[metaOffset+7] = byte(ptr >> 24)

		// len
		chunkLen := int32(len(chunk))
		memData[metaOffset+8] = byte(chunkLen)
		memData[metaOffset+9] = byte(chunkLen >> 8)
		memData[metaOffset+10] = byte(chunkLen >> 16)
		memData[metaOffset+11] = byte(chunkLen >> 24)

		// isLast
		if isLast {
			memData[metaOffset+12] = 1
		}

		logrus.WithFields(logrus.Fields{"chunkID": chunkID, "ptr": ptr, "len": chunkLen, "isLast": isLast}).Debug("[Streaming] Input chunk")
		return int32(metaOffset)
	}); err != nil {
		return err
	}

	// streaming_set_output_ready - WASM signals output is ready at given location
	if err := linker.DefineFunc(store, "env", "streaming_set_output_ready", func(caller *wasmtime.Caller, chunkID int32, ptr int32, chunkLen int32) int32 {
		if streamingState == nil {
			return 0 // Stub
		}

		// Get memory to read the output data
		mem := caller.GetExport("memory")
		if mem == nil {
			return -1 // Error
		}
		memory := mem.Memory()
		if memory == nil {
			return -1
		}
		memData := memory.UnsafeData(store)

		// Read the output chunk from memory
		if int(ptr)+int(chunkLen) > len(memData) {
			return -1 // Out of bounds
		}

		chunk := make([]byte, chunkLen)
		copy(chunk, memData[ptr:ptr+chunkLen])

		// Add to streaming state
		streamingState.AddOutputChunk(chunkID, chunk)

		logrus.WithFields(logrus.Fields{"chunkID": chunkID, "len": chunkLen}).Debug("[Streaming] Output chunk ready")
		return 0 // Success
	}); err != nil {
		return err
	}

	// streaming_get_next_output_ptr - Returns pointer where WASM should write next output chunk
	if err := linker.DefineFunc(store, "env", "streaming_get_next_output_ptr", func(caller *wasmtime.Caller) int32 {
		if streamingState == nil {
			return 0 // Stub
		}

		// Calculate next output chunk location
		nextChunkID := streamingState.GetOutputCount()
		nextPtr := streamingState.GetOutputChunkPtr(nextChunkID)

		logrus.WithFields(logrus.Fields{"chunkID": nextChunkID, "ptr": nextPtr}).Debug("[Streaming] Next output ptr")
		return nextPtr
	}); err != nil {
		return err
	}

	// streaming_chunk_read - Read a chunk's data into the provided buffer (host -> WASM)
	if err := linker.DefineFunc(store, "env", "streaming_chunk_read", func(caller *wasmtime.Caller, chunkID int32, destPtr int32, maxLen int32) int32 {
		if streamingState == nil {
			return -1 // Stub
		}

		// Get memory
		mem := caller.GetExport("memory")
		if mem == nil {
			return -1
		}
		memory := mem.Memory()
		if memory == nil {
			return -1
		}
		memData := memory.UnsafeData(store)

		// Read chunk data
		bytesRead := streamingState.ReadChunkInto(chunkID, destPtr, maxLen, memData)
		if bytesRead > 0 {
			logrus.WithFields(logrus.Fields{"chunkID": chunkID, "bytesRead": bytesRead, "destPtr": destPtr}).Debug("[Streaming] Read chunk")
		}
		return bytesRead
	}); err != nil {
		return err
	}

	return nil
}

// DefineFunctionFlyPythonBridge registers env.ff_* host functions that bridge
// Python code running in MicroPython to FunctionFly platform services.
// These functions are called by the _functionfly Python module via the
// shared memory + mp_js_hook protocol.
func DefineFunctionFlyPythonBridge(linker *wasmtime.Linker, store *wasmtime.Store, handler HostFunctionHandler) error {
	// env.ff_log(level, msg_ptr, msg_len)
	if err := linker.DefineFunc(store, "env", "ff_log", func(caller *wasmtime.Caller, level, msgPtr, msgLen int32) {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return
		}
		memoryData := memory.UnsafeData(store)
		if msgPtr < 0 || int(msgPtr)+int(msgLen) > len(memoryData) {
			return
		}
		message := string(memoryData[msgPtr : msgPtr+msgLen])
		handler.Log(message)
	}); err != nil {
		return err
	}

	// env.ff_get_env(name_ptr, name_len, val_ptr, val_len_ptr) -> i32
	if err := linker.DefineFunc(store, "env", "ff_get_env", func(caller *wasmtime.Caller, namePtr, nameLen, valPtr, valLenPtr int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, namePtr, nameLen); err != nil {
			return -2
		}
		name := string(memoryData[namePtr : namePtr+nameLen])
		value, err := handler.GetEnv(name)
		if err != nil {
			return -1
		}
		valueBytes := []byte(value)
		valLen := len(valueBytes)
		if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
			return -2
		}
		copy(memoryData[valPtr:], valueBytes)
		if err := validateMemoryBounds(memoryData, valLenPtr, 4); err != nil {
			return -2
		}
		memoryData[valLenPtr] = byte(valLen)
		memoryData[valLenPtr+1] = byte(valLen >> 8)
		memoryData[valLenPtr+2] = byte(valLen >> 16)
		memoryData[valLenPtr+3] = byte(valLen >> 24)
		return 0
	}); err != nil {
		return err
	}

	// env.ff_kv_get(key_ptr, key_len, val_ptr, val_len_ptr) -> i32
	if err := linker.DefineFunc(store, "env", "ff_kv_get", func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLenPtr int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, keyPtr, keyLen); err != nil {
			return -2
		}
		key := string(memoryData[keyPtr : keyPtr+keyLen])
		value, err := handler.KVGet(key)
		if err != nil {
			return -1
		}
		valueBytes := []byte(value)
		valLen := len(valueBytes)
		if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
			return -2
		}
		copy(memoryData[valPtr:], valueBytes)
		if err := validateMemoryBounds(memoryData, valLenPtr, 4); err != nil {
			return -2
		}
		memoryData[valLenPtr] = byte(valLen)
		memoryData[valLenPtr+1] = byte(valLen >> 8)
		memoryData[valLenPtr+2] = byte(valLen >> 16)
		memoryData[valLenPtr+3] = byte(valLen >> 24)
		return 0
	}); err != nil {
		return err
	}

	// env.ff_kv_set(key_ptr, key_len, val_ptr, val_len) -> i32
	if err := linker.DefineFunc(store, "env", "ff_kv_set", func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLen int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, keyPtr, keyLen); err != nil {
			return -2
		}
		key := string(memoryData[keyPtr : keyPtr+keyLen])
		if err := validateMemoryBounds(memoryData, valPtr, valLen); err != nil {
			return -2
		}
		value := string(memoryData[valPtr : valPtr+valLen])
		if err := handler.KVSet(key, value); err != nil {
			return -1
		}
		return 0
	}); err != nil {
		return err
	}

	// env.ff_state_get(path_ptr, path_len, val_ptr, val_len_ptr) -> i32
	if err := linker.DefineFunc(store, "env", "ff_state_get", func(caller *wasmtime.Caller, pathPtr, pathLen, valPtr, valLenPtr int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
			return -2
		}
		path := string(memoryData[pathPtr : pathPtr+pathLen])
		value, err := handler.StateGet(path)
		if err != nil {
			return -1
		}
		valueBytes := []byte(value)
		valLen := len(valueBytes)
		if globalSecurityConfig != nil && uint32(valLen) > globalSecurityConfig.MaxOutputSize {
			return -1
		}
		if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
			return -2
		}
		copy(memoryData[valPtr:], valueBytes)
		if err := validateMemoryBounds(memoryData, valLenPtr, 4); err != nil {
			return -2
		}
		memoryData[valLenPtr] = byte(valLen)
		memoryData[valLenPtr+1] = byte(valLen >> 8)
		memoryData[valLenPtr+2] = byte(valLen >> 16)
		memoryData[valLenPtr+3] = byte(valLen >> 24)
		return 0
	}); err != nil {
		return err
	}

	// env.ff_state_set(path_ptr, path_len, val_ptr, val_len) -> i32
	if err := linker.DefineFunc(store, "env", "ff_state_set", func(caller *wasmtime.Caller, pathPtr, pathLen, valPtr, valLen int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
			return -2
		}
		path := string(memoryData[pathPtr : pathPtr+pathLen])
		if err := validateMemoryBounds(memoryData, valPtr, valLen); err != nil {
			return -2
		}
		value := string(memoryData[valPtr : valPtr+valLen])
		if globalSecurityConfig != nil && uint32(len(value)) > globalSecurityConfig.MaxInputSize {
			return -1
		}
		if err := handler.StateSet(path, value); err != nil {
			return -1
		}
		return 0
	}); err != nil {
		return err
	}

	// env.ff_state_delete(path_ptr, path_len) -> i32
	if err := linker.DefineFunc(store, "env", "ff_state_delete", func(caller *wasmtime.Caller, pathPtr, pathLen int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
			return -2
		}
		path := string(memoryData[pathPtr : pathPtr+pathLen])
		if err := handler.StateDelete(path); err != nil {
			return -1
		}
		return 0
	}); err != nil {
		return err
	}

	// env.ff_state_get_fabric(fabric_id_ptr, fabric_id_len, resp_ptr, resp_len_ptr) -> i32
	if err := linker.DefineFunc(store, "env", "ff_state_get_fabric", func(caller *wasmtime.Caller, fabricIDPtr, fabricIDLen, respPtr, respLenPtr int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, fabricIDPtr, fabricIDLen); err != nil {
			return -2
		}
		fabricID := string(memoryData[fabricIDPtr : fabricIDPtr+fabricIDLen])
		result, err := handler.StateGetFabric(fabricID)
		if err != nil {
			return -1
		}
		resultBytes := []byte(result)
		respLen := len(resultBytes)
		if globalSecurityConfig != nil && uint32(respLen) > globalSecurityConfig.MaxOutputSize {
			return -1
		}
		if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
			return -2
		}
		copy(memoryData[respPtr:], resultBytes)
		if err := validateMemoryBounds(memoryData, respLenPtr, 4); err != nil {
			return -2
		}
		memoryData[respLenPtr] = byte(respLen)
		memoryData[respLenPtr+1] = byte(respLen >> 8)
		memoryData[respLenPtr+2] = byte(respLen >> 16)
		memoryData[respLenPtr+3] = byte(respLen >> 24)
		return 0
	}); err != nil {
		return err
	}

	// env.ff_state_create_snapshot(path_ptr, path_len, label_ptr, label_len, resp_ptr, resp_len_ptr) -> i32
	if err := linker.DefineFunc(store, "env", "ff_state_create_snapshot", func(caller *wasmtime.Caller, pathPtr, pathLen, labelPtr, labelLen, respPtr, respLenPtr int32) int32 {
		memory := caller.GetExport("memory").Memory()
		if memory == nil {
			return -2
		}
		memoryData := memory.UnsafeData(store)
		if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
			return -2
		}
		path := string(memoryData[pathPtr : pathPtr+pathLen])
		label := ""
		if labelLen > 0 {
			if err := validateMemoryBounds(memoryData, labelPtr, labelLen); err != nil {
				return -2
			}
			label = string(memoryData[labelPtr : labelPtr+labelLen])
		}
		result, err := handler.StateCreateSnapshot(path, label)
		if err != nil {
			return -1
		}
		resultBytes := []byte(result)
		respLen := len(resultBytes)
		if globalSecurityConfig != nil && uint32(respLen) > globalSecurityConfig.MaxOutputSize {
			return -1
		}
		if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
			return -2
		}
		copy(memoryData[respPtr:], resultBytes)
		if err := validateMemoryBounds(memoryData, respLenPtr, 4); err != nil {
			return -2
		}
		memoryData[respLenPtr] = byte(respLen)
		memoryData[respLenPtr+1] = byte(respLen >> 8)
		memoryData[respLenPtr+2] = byte(respLen >> 16)
		memoryData[respLenPtr+3] = byte(respLen >> 24)
		return 0
	}); err != nil {
		return err
	}

	return nil
}
