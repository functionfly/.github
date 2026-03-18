//go:build cgo

package wasm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bytecodealliance/wasmtime-go/v19"
)

// Global security config for host functions (set during runtime initialization)
var globalSecurityConfig *WASMSecurityConfig

// SetGlobalSecurityConfig sets the global security configuration for host functions
func SetGlobalSecurityConfig(config *WASMSecurityConfig) {
	globalSecurityConfig = config
}

// validateMemoryBounds validates pointer and size are within memory bounds
func validateMemoryBounds(memoryData []byte, ptr, size int32) error {
	if ptr < 0 {
		return fmt.Errorf("negative pointer: %d", ptr)
	}
	if size < 0 {
		return fmt.Errorf("negative size: %d", size)
	}
	if int(ptr)+int(size) > len(memoryData) {
		return fmt.Errorf("memory access out of bounds: ptr=%d size=%d memory=%d",
			ptr, size, len(memoryData))
	}
	return nil
}

// validateDomain checks if a URL's domain is in the allowed list
func validateDomain(requestURL string) error {
	if globalSecurityConfig == nil || len(globalSecurityConfig.AllowedDomains) == 0 {
		return nil // No restrictions
	}

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(parsedURL.Hostname())
	for _, allowed := range globalSecurityConfig.AllowedDomains {
		allowedLower := strings.ToLower(allowed)
		if host == allowedLower || strings.HasSuffix(host, "."+allowedLower) {
			return nil
		}
	}

	return fmt.Errorf("domain not allowed: %s (allowed: %v)", host, globalSecurityConfig.AllowedDomains)
}

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

	// Define fetch function with domain allowlist
	// (param $req_ptr i32) (param $req_len i32) (param $resp_ptr i32) (param $resp_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "fetch",
		func(caller *wasmtime.Caller, reqPtr, reqLen, respPtr, respLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				log.Printf("WASM fetch: memory not found")
				return -1 // error
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds
			if err := validateMemoryBounds(memoryData, reqPtr, reqLen); err != nil {
				log.Printf("WASM fetch: %v", err)
				return -1
			}

			requestStr := string(memoryData[reqPtr : reqPtr+reqLen])

			// Parse request to validate domain
			var fetchReq FetchRequest
			if err := json.Unmarshal([]byte(requestStr), &fetchReq); err != nil {
				log.Printf("WASM fetch: failed to parse request: %v", err)
				return -1
			}

			// Security: Domain allowlist check
			if err := validateDomain(fetchReq.URL); err != nil {
				log.Printf("WASM fetch: %v", err)
				return -1
			}

			// Perform fetch
			response, err := handler.Fetch(requestStr)
			if err != nil {
				return -1
			}

			responseBytes := []byte(response)
			respLen := len(responseBytes)

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(respLen) > globalSecurityConfig.MaxOutputSize {
				log.Printf("WASM fetch: response too large: %d > %d", respLen, globalSecurityConfig.MaxOutputSize)
				return -1
			}

			// Security: Validate response pointer bounds
			if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
				log.Printf("WASM fetch: %v", err)
				return -1
			}

			// Write response
			copy(memoryData[respPtr:], responseBytes)

			// Write response length
			if err := validateMemoryBounds(memoryData, respLenPtr, 4); err != nil {
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

	// Define kv_get function with security validation
	// (param $key_ptr i32) (param $key_len i32) (param $val_ptr i32) (param $val_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "kv_get",
		func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds
			if err := validateMemoryBounds(memoryData, keyPtr, keyLen); err != nil {
				log.Printf("WASM kv_get: %v", err)
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

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(valLen) > globalSecurityConfig.MaxOutputSize {
				log.Printf("WASM kv_get: value too large: %d > %d", valLen, globalSecurityConfig.MaxOutputSize)
				return -1
			}

			// Security: Validate pointer bounds
			if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
				log.Printf("WASM kv_get: %v", err)
				return -1
			}

			// Write value
			copy(memoryData[valPtr:], valueBytes)

			// Write value length
			if err := validateMemoryBounds(memoryData, valLenPtr, 4); err != nil {
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

	// Define kv_set function with security validation
	// (param $key_ptr i32) (param $key_len i32) (param $val_ptr i32) (param $val_len i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "kv_set",
		func(caller *wasmtime.Caller, keyPtr, keyLen, valPtr, valLen int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for key
			if err := validateMemoryBounds(memoryData, keyPtr, keyLen); err != nil {
				log.Printf("WASM kv_set: %v", err)
				return -1
			}

			key := string(memoryData[keyPtr : keyPtr+keyLen])

			// Security: Validate pointer bounds for value
			if err := validateMemoryBounds(memoryData, valPtr, valLen); err != nil {
				log.Printf("WASM kv_set: %v", err)
				return -1
			}

			value := string(memoryData[valPtr : valPtr+valLen])

			// Security: Check input size limit
			if globalSecurityConfig != nil && uint32(len(value)) > globalSecurityConfig.MaxInputSize {
				log.Printf("WASM kv_set: value too large: %d > %d", len(value), globalSecurityConfig.MaxInputSize)
				return -1
			}

			// Set value
			if err := handler.KVSet(key, value); err != nil {
				return -1
			}

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define kv_set function: %w", err)
	}

	// Define get_env function with security validation
	// (param $name_ptr i32) (param $name_len i32) (param $val_ptr i32) (param $val_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "get_env",
		func(caller *wasmtime.Caller, namePtr, nameLen, valPtr, valLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds
			if err := validateMemoryBounds(memoryData, namePtr, nameLen); err != nil {
				log.Printf("WASM get_env: %v", err)
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

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(valLen) > globalSecurityConfig.MaxOutputSize {
				log.Printf("WASM get_env: value too large: %d > %d", valLen, globalSecurityConfig.MaxOutputSize)
				return -1
			}

			// Security: Validate pointer bounds
			if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
				log.Printf("WASM get_env: %v", err)
				return -1
			}

			// Write value
			copy(memoryData[valPtr:], valueBytes)

			// Write value length
			if err := validateMemoryBounds(memoryData, valLenPtr, 4); err != nil {
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

