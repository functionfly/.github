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

// defineStateFabricHostFunctions registers StateFabric host functions with the linker
func defineStateFabricHostFunctions(linker *wasmtime.Linker, store *wasmtime.Store, handler HostFunctionHandler) error {
	// Define state_get function
	// (param $path_ptr i32) (param $path_len i32) (param $val_ptr i32) (param $val_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "state_get",
		func(caller *wasmtime.Caller, pathPtr, pathLen, valPtr, valLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for path
			if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
				return -1
			}

			path := string(memoryData[pathPtr : pathPtr+pathLen])

			// Get value from StateFabric
			value, err := handler.StateGet(path)
			if err != nil {
				return -1
			}

			valueBytes := []byte(value)
			valLen := len(valueBytes)

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(valLen) > globalSecurityConfig.MaxOutputSize {
				return -1
			}

			// Security: Validate pointer bounds for value
			if err := validateMemoryBounds(memoryData, valPtr, int32(valLen)); err != nil {
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
		return fmt.Errorf("failed to define state_get function: %w", err)
	}

	// Define state_set function
	// (param $path_ptr i32) (param $path_len i32) (param $val_ptr i32) (param $val_len i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "state_set",
		func(caller *wasmtime.Caller, pathPtr, pathLen, valPtr, valLen int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for path
			if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
				return -1
			}

			path := string(memoryData[pathPtr : pathPtr+pathLen])

			// Security: Validate pointer bounds for value
			if err := validateMemoryBounds(memoryData, valPtr, valLen); err != nil {
				return -1
			}

			value := string(memoryData[valPtr : valPtr+valLen])

			// Security: Check input size limit
			if globalSecurityConfig != nil && uint32(len(value)) > globalSecurityConfig.MaxInputSize {
				return -1
			}

			// Set value in StateFabric
			if err := handler.StateSet(path, value); err != nil {
				return -1
			}

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define state_set function: %w", err)
	}

	// Define state_delete function
	// (param $path_ptr i32) (param $path_len i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "state_delete",
		func(caller *wasmtime.Caller, pathPtr, pathLen int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for path
			if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
				return -1
			}

			path := string(memoryData[pathPtr : pathPtr+pathLen])

			// Delete value from StateFabric
			if err := handler.StateDelete(path); err != nil {
				return -1
			}

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define state_delete function: %w", err)
	}

	// Define state_get_fabric function
	// (param $fabric_id_ptr i32) (param $fabric_id_len i32) (param $resp_ptr i32) (param $resp_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "state_get_fabric",
		func(caller *wasmtime.Caller, fabricIDPtr, fabricIDLen, respPtr, respLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for fabric ID
			if err := validateMemoryBounds(memoryData, fabricIDPtr, fabricIDLen); err != nil {
				return -1
			}

			fabricID := string(memoryData[fabricIDPtr : fabricIDPtr+fabricIDLen])

			// Get fabric metadata from StateFabric
			fabricInfo, err := handler.StateGetFabric(fabricID)
			if err != nil {
				return -1
			}

			fabricBytes := []byte(fabricInfo)
			respLen := len(fabricBytes)

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(respLen) > globalSecurityConfig.MaxOutputSize {
				return -1
			}

			// Security: Validate pointer bounds for response
			if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
				return -1
			}

			// Write fabric info
			copy(memoryData[respPtr:], fabricBytes)

			// Write response length
			if err := validateMemoryBounds(memoryData, respLenPtr, 4); err != nil {
				return -1
			}
			memoryData[respLenPtr] = byte(respLen)
			memoryData[respLenPtr+1] = byte(respLen >> 8)
			memoryData[respLenPtr+2] = byte(respLen >> 16)
			memoryData[respLenPtr+3] = byte(respLen >> 24)

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define state_get_fabric function: %w", err)
	}

	// Define state_create_snapshot function
	// (param $path_ptr i32) (param $path_len i32) (param $label_ptr i32) (param $label_len i32) (param $resp_ptr i32) (param $resp_len_ptr i32) (result i32)
	if err := linker.DefineFunc(store, "functionfly", "state_create_snapshot",
		func(caller *wasmtime.Caller, pathPtr, pathLen, labelPtr, labelLen, respPtr, respLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Validate pointer bounds for path
			if err := validateMemoryBounds(memoryData, pathPtr, pathLen); err != nil {
				return -1
			}

			path := string(memoryData[pathPtr : pathPtr+pathLen])

			// Validate label pointer (can be empty)
			var label string
			if labelLen > 0 {
				if err := validateMemoryBounds(memoryData, labelPtr, labelLen); err != nil {
					return -1
				}
				label = string(memoryData[labelPtr : labelPtr+labelLen])
			}

			// Create snapshot via StateFabric
			snapshot, err := handler.StateCreateSnapshot(path, label)
			if err != nil {
				return -1
			}

			snapshotBytes := []byte(snapshot)
			respLen := len(snapshotBytes)

			// Security: Check output size limit
			if globalSecurityConfig != nil && uint32(respLen) > globalSecurityConfig.MaxOutputSize {
				return -1
			}

			// Security: Validate pointer bounds for response
			if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
				return -1
			}

			// Write snapshot info
			copy(memoryData[respPtr:], snapshotBytes)

			// Write response length
			if err := validateMemoryBounds(memoryData, respLenPtr, 4); err != nil {
				return -1
			}
			memoryData[respLenPtr] = byte(respLen)
			memoryData[respLenPtr+1] = byte(respLen >> 8)
			memoryData[respLenPtr+2] = byte(respLen >> 16)
			memoryData[respLenPtr+3] = byte(respLen >> 24)

			return 0
		}); err != nil {
		return fmt.Errorf("failed to define state_create_snapshot function: %w", err)
	}

	return nil
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
			value := handler.GetEnv(name)

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

	// Define ai_infer function for AI inference via AI Gateway
	// (param $model_ptr i32) (param $model_len i32)
	// (param $input_ptr i32) (param $input_len i32)
	// (param $params_ptr i32) (param $params_len i32)
	// (param $resp_ptr i32) (param $resp_len_ptr i32)
	// (result i32) - 0 = success, -1 = error
	if err := linker.DefineFunc(store, "functionfly", "ai_infer",
		func(caller *wasmtime.Caller, modelPtr, modelLen, inputPtr, inputLen, paramsPtr, paramsLen, respPtr, respLenPtr int32) int32 {
			memory := caller.GetExport("memory").Memory()
			if memory == nil {
				log.Printf("WASM ai_infer: memory not found")
				return -1
			}

			memoryData := memory.UnsafeData(store)

			// Security: Check if AI inference is enabled
			if globalSecurityConfig == nil || !globalSecurityConfig.AIInference.Enabled {
				log.Printf("WASM ai_infer: ai inference not enabled")
				return -1
			}

			// Security: Validate model pointer bounds
			if err := validateMemoryBounds(memoryData, modelPtr, modelLen); err != nil {
				log.Printf("WASM ai_infer: invalid model pointer: %v", err)
				return -1
			}

			// Security: Validate input pointer bounds
			if err := validateMemoryBounds(memoryData, inputPtr, inputLen); err != nil {
				log.Printf("WASM ai_infer: invalid input pointer: %v", err)
				return -1
			}

			// Security: Validate params pointer bounds (params can be empty/zero)
			if paramsLen > 0 {
				if err := validateMemoryBounds(memoryData, paramsPtr, paramsLen); err != nil {
					log.Printf("WASM ai_infer: invalid params pointer: %v", err)
					return -1
				}
			}

			// Read model name
			model := string(memoryData[modelPtr : modelPtr+modelLen])

			// Read input data
			input := make([]byte, inputLen)
			copy(input, memoryData[inputPtr:inputPtr+inputLen])

			// Read params (optional)
			var params string
			if paramsLen > 0 {
				params = string(memoryData[paramsPtr : paramsPtr+paramsLen])
			}

			// Perform AI inference
			response, err := handler.AIInference(model, input, params)
			if err != nil {
				log.Printf("WASM ai_infer: inference failed: %v", err)
				return -1
			}

			responseBytes := []byte(response)
			respLen := len(responseBytes)

			// Security: Check output size limit against AI inference config
			maxOutputSize := uint32(globalSecurityConfig.AIInference.MaxModelSizeMB) * 1024 * 1024
			if uint32(respLen) > maxOutputSize {
				log.Printf("WASM ai_infer: response too large: %d > %d", respLen, maxOutputSize)
				return -1
			}

			// Security: Validate response pointer bounds
			if err := validateMemoryBounds(memoryData, respPtr, int32(respLen)); err != nil {
				log.Printf("WASM ai_infer: %v", err)
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
		return fmt.Errorf("failed to define ai_infer function: %w", err)
	}

	// Register StateFabric host functions for edge state access
	if err := defineStateFabricHostFunctions(linker, store, handler); err != nil {
		return fmt.Errorf("failed to define state fabric host functions: %w", err)
	}

	return nil
}
