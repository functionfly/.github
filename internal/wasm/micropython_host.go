//go:build cgo

package wasm

import (
	"log"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

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
		log.Printf("[Streaming] WASM acknowledged chunk %d (%d bytes), isLast=%v", chunkID, len, isLast != 0)
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

		log.Printf("[Streaming] Output chunk %d: ptr=%d len=%d", chunkID, ptr, chunkLen)
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

		log.Printf("[Streaming] Input chunk %d: ptr=%d len=%d isLast=%v", chunkID, ptr, chunkLen, isLast)
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

		log.Printf("[Streaming] Output chunk %d ready: %d bytes", chunkID, chunkLen)
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

		log.Printf("[Streaming] Next output ptr: chunkID=%d ptr=%d", nextChunkID, nextPtr)
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
			log.Printf("[Streaming] Read chunk %d: %d bytes into ptr %d", chunkID, bytesRead, destPtr)
		}
		return bytesRead
	}); err != nil {
		return err
	}

	return nil
}
