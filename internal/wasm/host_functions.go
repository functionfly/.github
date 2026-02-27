//go:build cgo

package wasm

import (
	"fmt"
	"log"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// defineHostFunctions registers all FunctionFly host functions with the linker
func defineHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store, handler HostFunctionHandler) error {

	// Define log function
	// (param $msg_ptr i32) (param $msg_len i32)
	if err := linker.DefineFunc(store, "functionfly", "log",
		func(caller *wasmtime.Caller, msgPtr, msgLen int32) {
			// Read message from memory
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				log.Printf("WASM log: memory not found")
				return
			}

			memoryData := memory.UnsafeData(store)
			if msgPtr < 0 || int(msgPtr)+int(msgLen) > len(memoryData) {
				log.Printf("WASM log: invalid memory bounds")
				return
			}

			message := string(memoryData[msgPtr : msgPtr+msgLen])
			handler.Log(message)
		}); err != nil {
		return fmt.Errorf("failed to define log function: %w", err)
	}

	// Define fetch function
	// (param $req_ptr i32) (param $req_len i32) (param $resp_ptr i32) (param $resp_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "fetch",
		func(caller *wasmtime.Caller, reqPtr, reqLen, respPtr, respLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1 // error
			}

			memoryData := memory.UnsafeData(store)

			// Read request
			if reqPtr < 0 || int(reqPtr)+int(reqLen) > len(memoryData) {
				return -1
			}
			request := string(memoryData[reqPtr : reqPtr+reqLen])

			// Perform fetch
			response, err := handler.Fetch(request)
			if err != nil {
				return -1
			}

			responseBytes := []byte(response)
			respLen := len(responseBytes)

			// Check if response buffer is large enough
			if respPtr < 0 || int(respPtr)+respLen > len(memoryData) {
				return -1
			}

			// Write response
			copy(memoryData[respPtr:], responseBytes)

			// Write response length
			if respLenPtr < 0 || int(respLenPtr)+4 > len(memoryData) {
				return -1
			}
			memoryData[respLenPtr] = byte(respLen)
			memoryData[respLenPtr+1] = byte(respLen >> 8)
			memoryData[respLenPtr+2] = byte(respLen >> 16)
			memoryData[respLenPtr+3] = byte(respLen >> 24)

			return 0 // success
		}); err != nil {
		return fmt.Errorf("failed to define fetch function: %w", err)
	}

	// Define kv_get function
	// (param $key_ptr i32) (param $key_len i32) (param $val_ptr i32) (param $val_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "kv_get",
		func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Read key
			if keyPtr < 0 || int(keyPtr)+int(keyLen) > len(memoryData) {
				return -1
			}
			key := string(memoryData[keyPtr : keyPtr+keyLen])

			// Get value
			value, err := handler.KVGet(key)
			if err != nil {
				return -1
			}

			valueBytes := []byte(value)
			valLen := len(valueBytes)

			// Check if value buffer is large enough
			if valPtr < 0 || int(valPtr)+valLen > len(memoryData) {
				return -1
			}

			// Write value
			copy(memoryData[valPtr:], valueBytes)

			// Write value length
			if valLenPtr < 0 || int(valLenPtr)+4 > len(memoryData) {
				return -1
			}
			memoryData[valLenPtr] = byte(valLen)
			memoryData[valLenPtr+1] = byte(valLen >> 8)
			memoryData[valLenPtr+2] = byte(valLen >> 16)
			memoryData[valLenPtr+3] = byte(valLen >> 24)

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define kv_get function: %w", err)
	}

	// Define kv_set function
	// (param $key_ptr i32) (param $key_len i32) (param $val_ptr i32) (param $val_len i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "kv_set",
		func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLen int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Read key
			if keyPtr < 0 || int(keyPtr)+int(keyLen) > len(memoryData) {
				return -1
			}
			key := string(memoryData[keyPtr : keyPtr+keyLen])

			// Read value
			if valPtr < 0 || int(valPtr)+int(valLen) > len(memoryData) {
				return -1
			}
			value := string(memoryData[valPtr : valPtr+valLen])

			// Set value
			if err := handler.KVSet(key, value); err != nil {
				return -1
			}

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define kv_set function: %w", err)
	}

	// Define get_env function
	// (param $name_ptr i32) (param $name_len i32) (param $val_ptr i32) (param $val_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "get_env",
		func(caller *wasmtime.Caller, namePtr, nameLen, valPtr, valLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Read name
			if namePtr < 0 || int(namePtr)+int(nameLen) > len(memoryData) {
				return -1
			}
			name := string(memoryData[namePtr : namePtr+nameLen])

			// Get environment variable
			value, err := handler.GetEnv(name)
			if err != nil {
				return -1
			}

			valueBytes := []byte(value)
			valLen := len(valueBytes)

			// Check if value buffer is large enough
			if valPtr < 0 || int(valPtr)+valLen > len(memoryData) {
				return -1
			}

			// Write value
			copy(memoryData[valPtr:], valueBytes)

			// Write value length
			if valLenPtr < 0 || int(valLenPtr)+4 > len(memoryData) {
				return -1
			}
			memoryData[valLenPtr] = byte(valLen)
			memoryData[valLenPtr+1] = byte(valLen >> 8)
			memoryData[valLenPtr+2] = byte(valLen >> 16)
			memoryData[valLenPtr+3] = byte(valLen >> 24)

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define get_env function: %w", err)
	}

	return nil
}
