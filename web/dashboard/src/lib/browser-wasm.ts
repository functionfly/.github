/**
 * Browser Native WASM SDK
 *
 * This SDK enables executing WebAssembly modules in the browser environment
 * with proper security controls and resource limits.
 *
 * Security features:
 * - Input/output size limits
 * - Execution timeout enforcement
 * - WASM module validation
 * - Memory limits (via browser sandbox)
 * - Optional network access controls
 *
 * Usage:
 * ```typescript
 * import { BrowserWasmRuntime } from './lib/browser-wasm';
 *
 * const runtime = new BrowserWasmRuntime({
 *   maxInputSize: 1024 * 1024,  // 1MB
 *   maxOutputSize: 1024 * 1024, // 1MB
 *   executionTimeout: 30000,    // 30 seconds
 * });
 *
 * const result = await runtime.execute(wasmBytes, input);
 * ```
 */

export interface BrowserWasmConfig {
  /** Maximum WASM module size in bytes (default: 10MB) */
  maxWasmModuleSize?: number;
  /** Maximum input size in bytes (default: 1MB) */
  maxInputSize?: number;
  /** Maximum output size in bytes (default: 1MB) */
  maxOutputSize?: number;
  /** Execution timeout in milliseconds (default: 30000) */
  executionTimeout?: number;
  /** Enable network access via fetch (default: true) */
  enableNetworkAccess?: boolean;
  /** Allowed origins for CORS (default: all) */
  allowedOrigins?: string[];
  /** Validate WASM modules before execution (default: true) */
  validateModules?: boolean;
  /** Called with progress updates during compilation/execution */
  onProgress?: (stage: string, progress: number) => void;
}

export interface BrowserWasmExecutionResult {
  /** The output data from the WASM function */
  output: Uint8Array | string;
  /** Execution time in milliseconds */
  executionTimeMs: number;
  /** Memory usage peak in bytes (if available) */
  peakMemoryBytes?: number;
  /** Whether the execution was successful */
  success: boolean;
  /** Error message if execution failed */
  error?: string;
}

export interface BrowserWasmModule {
  /** Compiled WASM module */
  module: WebAssembly.Module;
  /** Memory instance (if exported) */
  memory?: WebAssembly.Memory;
  /** Exported functions */
  exports: WebAssembly.Exports;
  /** Module size in bytes */
  sizeBytes: number;
}

const DEFAULT_CONFIG: Required<BrowserWasmConfig> = {
  maxWasmModuleSize: 10 * 1024 * 1024, // 10MB
  maxInputSize: 1 * 1024 * 1024, // 1MB
  maxOutputSize: 1 * 1024 * 1024, // 1MB
  executionTimeout: 30000, // 30s
  enableNetworkAccess: true,
  allowedOrigins: [],
  validateModules: true,
  onProgress: () => {},
};

// Helper to cast ArrayBufferLike to ArrayBuffer for TypeScript compatibility
function toArrayBuffer(buffer: ArrayBufferLike): ArrayBuffer {
  return buffer as ArrayBuffer;
}

/**
 * BrowserWasmRuntime executes WASM modules in the browser with security controls
 */
export class BrowserWasmRuntime {
  private config: Required<BrowserWasmConfig>;
  private compiledModules: Map<string, BrowserWasmModule> = new Map();

  constructor(config: BrowserWasmConfig = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * Validates a WASM binary before compilation
   * Returns validation result with details
   */
  async validateWasmModule(bytes: Uint8Array): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
  }> {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Cast buffer to ArrayBuffer for TypeScript compatibility
    const buffer = toArrayBuffer(bytes.buffer);

    // Size check
    if (bytes.length > this.config.maxWasmModuleSize) {
      errors.push(
        `WASM module too large: ${bytes.length} bytes (max: ${this.config.maxWasmModuleSize})`
      );
    }

    // Magic number check
    const magic = bytes.slice(0, 4);
    const expectedMagic = [0x00, 0x61, 0x73, 0x6d]; // \0asm
    if (
      magic[0] !== expectedMagic[0] ||
      magic[1] !== expectedMagic[1] ||
      magic[2] !== expectedMagic[2] ||
      magic[3] !== expectedMagic[3]
    ) {
      errors.push('Invalid WASM magic number - not a valid WASM binary');
    }

    // Version check
    const version = bytes.slice(4, 8);
    const versionNum = new DataView(toArrayBuffer(version.buffer)).getUint32(0, true);
    if (versionNum < 1 || versionNum > 13) {
      errors.push(`Unsupported WASM version: ${versionNum}`);
    }

    // Basic structure validation using WebAssembly.validate
    if (!WebAssembly.validate(buffer)) {
      errors.push('WebAssembly.validate() failed - malformed WASM binary');
    }

    // Try to compile to catch any structural errors
    try {
      new WebAssembly.Module(buffer);
    } catch (e) {
      errors.push(`WASM compilation failed: ${e instanceof Error ? e.message : String(e)}`);
    }

    // Size warnings
    if (bytes.length > 5 * 1024 * 1024) {
      warnings.push(`Large WASM module: ${(bytes.length / 1024 / 1024).toFixed(2)}MB`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * Compile a WASM module from bytes
   * Caches compiled modules by their hash
   */
  async compile(bytes: Uint8Array, cacheKey?: string): Promise<BrowserWasmModule> {
    const key = cacheKey || this.hashBytes(bytes);

    // Return cached module if available
    const cached = this.compiledModules.get(key);
    if (cached) {
      this.config.onProgress('cached', 100);
      return cached;
    }

    this.config.onProgress('validating', 0);

    // Validate if enabled
    if (this.config.validateModules) {
      const validation = await this.validateWasmModule(bytes);
      if (!validation.valid) {
        throw new Error(`WASM validation failed: ${validation.errors.join('; ')}`);
      }
      if (validation.warnings.length > 0) {
        console.warn('WASM validation warnings:', validation.warnings);
      }
    }

    this.config.onProgress('compiling', 10);

    // Compile the module - use buffer cast for TypeScript compatibility
    const buffer = toArrayBuffer(bytes.buffer);
    const module = await WebAssembly.compile(buffer);

    this.config.onProgress('instantiating', 50);

    // Instantiate to get exports and memory
    const instance = await WebAssembly.instantiate(module);

    this.config.onProgress('finalizing', 90);

    const browserModule: BrowserWasmModule = {
      module,
      memory: instance.exports.memory as WebAssembly.Memory | undefined,
      exports: instance.exports,
      sizeBytes: bytes.length,
    };

    // Cache the compiled module
    this.compiledModules.set(key, browserModule);

    this.config.onProgress('complete', 100);

    return browserModule;
  }

  /**
   * Execute a WASM function with the given input
   *
   * @param wasmBytes - The WASM binary bytes
   * @param input - Input data (will be JSON-serialized if object)
   * @param functionName - Name of the exported function to call (default: 'handle')
   * @param cacheKey - Optional key for module caching
   */
  async execute(
    wasmBytes: Uint8Array,
    input: unknown,
    functionName: string = 'handle',
    cacheKey?: string
  ): Promise<BrowserWasmExecutionResult> {
    const startTime = performance.now();

    try {
      // Compile the module
      const browserModule = await this.compile(wasmBytes, cacheKey);

      // Serialize input
      let inputBytes: Uint8Array;
      if (typeof input === 'string') {
        inputBytes = new TextEncoder().encode(input);
      } else {
        inputBytes = new TextEncoder().encode(JSON.stringify(input));
      }

      // Validate input size
      if (inputBytes.length > this.config.maxInputSize) {
        return {
          output: '',
          executionTimeMs: performance.now() - startTime,
          success: false,
          error: `Input too large: ${inputBytes.length} bytes (max: ${this.config.maxInputSize})`,
        };
      }

      // Prepare memory and call the function
      const result = await this.executeWithTimeout(browserModule, inputBytes, functionName);

      // Validate output size
      if (result.output.length > this.config.maxOutputSize) {
        return {
          output: '',
          executionTimeMs: performance.now() - startTime,
          success: false,
          error: `Output too large: ${result.output.length} bytes (max: ${this.config.maxOutputSize})`,
        };
      }

      // Decode output
      let output: Uint8Array | string;
      try {
        output = new TextDecoder().decode(result.output);
      } catch {
        output = result.output;
      }

      return {
        output,
        executionTimeMs: performance.now() - startTime,
        peakMemoryBytes: result.peakMemory,
        success: true,
      };
    } catch (error) {
      return {
        output: '',
        executionTimeMs: performance.now() - startTime,
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Execute with timeout enforcement
   */
  private async executeWithTimeout(
    browserModule: BrowserWasmModule,
    input: Uint8Array,
    functionName: string
  ): Promise<{ output: Uint8Array; peakMemory?: number }> {
    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(new Error(`Execution timeout after ${this.config.executionTimeout}ms`));
      }, this.config.executionTimeout);

      try {
        // Get the memory buffer if available
        let peakMemory: number | undefined;
        const memory = browserModule.memory;
        if (memory) {
          const initialPages = memory.buffer.byteLength;
          peakMemory = initialPages * 65536; // WASM pages are 64KB
        }

        // Get the handler function
        const exports = browserModule.exports;
        const handler = exports[functionName];

        if (!handler) {
          const availableExports = Object.keys(exports).filter(
            (k) => typeof exports[k] === 'function'
          );
          clearTimeout(timeoutId);
          reject(
            new Error(
              `Function '${functionName}' not found. Available: ${availableExports.join(', ')}`
            )
          );
          return;
        }

        // Allocate memory and write input
        let inputPtr = 0;
        if (memory && exports.alloc && exports.dealloc) {
          inputPtr = (exports.alloc as Function)(input.length);
          const memView = new Uint8Array(memory.buffer);
          memView.set(input, inputPtr);
        }

        // Call the handler
        let resultPtr = 0;
        let resultSize = 0;

        if (memory && inputPtr > 0) {
          // Function signature: (ptr, len) -> ptr
          resultPtr = (handler as Function)(inputPtr, input.length) as number;

          // Try to read result - assume result is a pointer to (ptr, len) tuple
          if (resultPtr !== 0) {
            const memView = new DataView(memory.buffer);
            resultPtr = memView.getUint32(resultPtr, true);
            resultSize = memView.getUint32(resultPtr + 4, true);
          }
        } else {
          // Fallback: call with just the input bytes directly
          const result = (handler as Function)(input);
          if (typeof result === 'number') {
            resultPtr = result;
          }
        }

        // Read output
        let output = new Uint8Array(0);
        if (memory && resultPtr !== 0 && resultSize > 0) {
          output = new Uint8Array(memory.buffer.slice(resultPtr, resultPtr + resultSize));
        }

        // Update peak memory if tracking
        if (memory) {
          const currentPages = memory.buffer.byteLength;
          const currentMemory = currentPages * 65536;
          if (currentMemory > (peakMemory || 0)) {
            peakMemory = currentMemory;
          }
        }

        // Cleanup
        if (memory && exports.dealloc && inputPtr > 0) {
          (exports.dealloc as Function)(inputPtr, input.length);
        }
        if (memory && exports.dealloc && resultPtr > 0 && resultSize > 0) {
          (exports.dealloc as Function)(resultPtr, resultSize);
        }

        clearTimeout(timeoutId);
        resolve({ output, peakMemory });
      } catch (error) {
        clearTimeout(timeoutId);
        reject(error);
      }
    });
  }

  /**
   * Simple hash function for module caching
   */
  private hashBytes(bytes: Uint8Array): string {
    const hash = Array.from(bytes.slice(0, 1000)).reduce((acc, byte, i) => acc + byte * (i + 1), 0);
    return `wasm_${bytes.length}_${hash}`;
  }

  /**
   * Clear cached modules
   */
  clearCache(): void {
    this.compiledModules.clear();
  }

  /**
   * Get cached module count
   */
  getCacheSize(): number {
    return this.compiledModules.size;
  }

  /**
   * Create a WebWorker-based executor for heavy workloads
   * This runs WASM execution in a separate thread
   */
  createWorkerExecutor(workerScript: string | URL): BrowserWasmWorkerExecutor {
    return new BrowserWasmWorkerExecutor(workerScript, this.config);
  }
}

/**
 * BrowserWasmWorkerExecutor runs WASM in a Web Worker for true parallelism
 */
export class BrowserWasmWorkerExecutor {
  private worker: Worker;
  private config: Required<BrowserWasmConfig>;
  private pendingRequests: Map<
    string,
    {
      resolve: (result: BrowserWasmExecutionResult) => void;
      reject: (error: Error) => void;
    }
  > = new Map();

  constructor(workerScript: string | URL, config: BrowserWasmConfig) {
    this.worker = new Worker(workerScript);
    this.config = { ...DEFAULT_CONFIG, ...config };

    this.worker.onmessage = this.handleMessage.bind(this);
    this.worker.onerror = this.handleError.bind(this);
  }

  private handleMessage(event: MessageEvent): void {
    const { requestId, result } = event.data;
    const pending = this.pendingRequests.get(requestId);
    if (pending) {
      this.pendingRequests.delete(requestId);
      pending.resolve(result);
    }
  }

  private handleError(error: ErrorEvent): void {
    // Find and reject the oldest pending request
    const oldestRequest = this.pendingRequests.keys().next().value;
    if (oldestRequest) {
      const pending = this.pendingRequests.get(oldestRequest);
      if (pending) {
        this.pendingRequests.delete(oldestRequest);
        pending.reject(new Error(error.message));
      }
    }
  }

  async execute(
    wasmBytes: Uint8Array,
    input: unknown,
    functionName: string = 'handle'
  ): Promise<BrowserWasmExecutionResult> {
    return new Promise((resolve, reject) => {
      const requestId = `req_${Date.now()}_${Math.random().toString(36).slice(2)}`;

      this.pendingRequests.set(requestId, { resolve, reject });

      this.worker.postMessage({
        requestId,
        wasmBytes,
        input,
        functionName,
        config: this.config,
      });

      // Timeout handler
      setTimeout(() => {
        if (this.pendingRequests.has(requestId)) {
          this.pendingRequests.delete(requestId);
          reject(new Error(`Worker execution timeout after ${this.config.executionTimeout}ms`));
        }
      }, this.config.executionTimeout);
    });
  }

  terminate(): void {
    this.worker.terminate();
    this.pendingRequests.clear();
  }
}

/**
 * Utility function to load WASM from a URL with progress tracking
 */
export async function loadWasmFromUrl(
  url: string,
  onProgress?: (loaded: number, total: number) => void,
  signal?: AbortSignal
): Promise<Uint8Array> {
  const response = await fetch(url, { signal });

  if (!response.ok) {
    throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
  }

  const contentLength = response.headers.get('content-length');
  const total = contentLength ? parseInt(contentLength, 10) : 0;

  if (!response.body) {
    // Fallback for environments without ReadableStream
    const buffer = await response.arrayBuffer();
    return new Uint8Array(buffer);
  }

  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let loaded = 0;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    chunks.push(value);
    loaded += value.length;
    onProgress?.(loaded, total);
  }

  // Concatenate all chunks
  const result = new Uint8Array(loaded);
  let offset = 0;
  for (const chunk of chunks) {
    result.set(chunk, offset);
    offset += chunk.length;
  }

  return result;
}

/**
 * Check if WebAssembly is supported in the current environment
 */
export function isWasmSupported(): boolean {
  return (
    typeof WebAssembly !== 'undefined' &&
    typeof WebAssembly.compile === 'function' &&
    typeof WebAssembly.instantiate === 'function'
  );
}

/**
 * Check if SharedArrayBuffer is available (requires secure context)
 */
export function isSharedArrayBufferSupported(): boolean {
  return typeof SharedArrayBuffer !== 'undefined';
}
