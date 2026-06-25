/// FunctionFly Swift SDK — Host function bindings.
///
/// These are WASM imports that call back into the FunctionFly runtime.
/// They are available when running inside a FunctionFly WASM sandbox.
/// When running natively (tests), they return safe no-op defaults.
import Foundation

// ─── Host Function Imports ────────────────────────────────────────────────────
// Marked @_silgen_name for WASM; stubs provided for native builds.

#if arch(wasm32)

@_silgen_name("functionfly_log")
private func _ff_log(_ msgPtr: UnsafePointer<UInt8>, _ msgLen: Int32)

@_silgen_name("functionfly_fetch")
private func _ff_fetch(_ reqPtr: UnsafePointer<UInt8>, _ reqLen: Int32,
                       _ respPtr: UnsafeMutablePointer<UInt8>, _ respLenPtr: UnsafeMutablePointer<Int32>) -> Int32

@_silgen_name("functionfly_kv_get")
private func _ff_kv_get(_ keyPtr: UnsafePointer<UInt8>, _ keyLen: Int32,
                        _ valPtr: UnsafeMutablePointer<UInt8>, _ valLenPtr: UnsafeMutablePointer<Int32>) -> Int32

@_silgen_name("functionfly_kv_set")
private func _ff_kv_set(_ keyPtr: UnsafePointer<UInt8>, _ keyLen: Int32,
                        _ valPtr: UnsafePointer<UInt8>, _ valLen: Int32) -> Int32

@_silgen_name("functionfly_get_env")
private func _ff_get_env(_ namePtr: UnsafePointer<UInt8>, _ nameLen: Int32,
                         _ valPtr: UnsafeMutablePointer<UInt8>, _ valLenPtr: UnsafeMutablePointer<Int32>) -> Int32

@_silgen_name("functionfly_ai_infer")
private func _ff_ai_infer(_ modelPtr: UnsafePointer<UInt8>, _ modelLen: Int32,
                          _ inputPtr: UnsafePointer<UInt8>, _ inputLen: Int32,
                          _ paramsPtr: UnsafePointer<UInt8>?, _ paramsLen: Int32,
                          _ respPtr: UnsafeMutablePointer<UInt8>, _ respLenPtr: UnsafeMutablePointer<Int32>) -> Int32

#else

// Native stubs — host functions are unavailable outside WASM.
private func _ff_log(_ msgPtr: UnsafePointer<UInt8>, _ msgLen: Int32) {}

private func _ff_fetch(_ reqPtr: UnsafePointer<UInt8>, _ reqLen: Int32,
                       _ respPtr: UnsafeMutablePointer<UInt8>, _ respLenPtr: UnsafeMutablePointer<Int32>) -> Int32 { return -1 }

private func _ff_kv_get(_ keyPtr: UnsafePointer<UInt8>, _ keyLen: Int32,
                        _ valPtr: UnsafeMutablePointer<UInt8>, _ valLenPtr: UnsafeMutablePointer<Int32>) -> Int32 { return -1 }

private func _ff_kv_set(_ keyPtr: UnsafePointer<UInt8>, _ keyLen: Int32,
                        _ valPtr: UnsafePointer<UInt8>, _ valLen: Int32) -> Int32 { return -1 }

private func _ff_get_env(_ namePtr: UnsafePointer<UInt8>, _ nameLen: Int32,
                         _ valPtr: UnsafeMutablePointer<UInt8>, _ valLenPtr: UnsafeMutablePointer<Int32>) -> Int32 { return -1 }

private func _ff_ai_infer(_ modelPtr: UnsafePointer<UInt8>, _ modelLen: Int32,
                          _ inputPtr: UnsafePointer<UInt8>, _ inputLen: Int32,
                          _ paramsPtr: UnsafePointer<UInt8>?, _ paramsLen: Int32,
                          _ respPtr: UnsafeMutablePointer<UInt8>, _ respLenPtr: UnsafeMutablePointer<Int32>) -> Int32 { return -1 }

#endif

// ─── Public API ──────────────────────────────────────────────────────────────

/// Host functions available to Swift functions running inside FunctionFly.
public enum HostFunctions {

    /// Log a message to the FunctionFly logging system.
    public static func log(_ message: String) {
        var msg = message
        msg.withUTF8 { buf in
            guard let base = buf.baseAddress else { return }
            _ff_log(base, Int32(buf.count))
        }
    }

    /// Perform an HTTP fetch request.
    ///
    /// - Parameter request: JSON string with URL, method, headers, body.
    /// - Returns: JSON response string, or nil on error.
    public static func fetch(_ request: String) -> String? {
        var req = request
        return req.withUTF8 { reqBuf in
            guard let reqPtr = reqBuf.baseAddress else { return nil }
            return callHostRead { respPtr, respLenPtr in
                _ff_fetch(reqPtr, Int32(reqBuf.count), respPtr, respLenPtr)
            }
        }
    }

    /// Convenience: GET request to a URL.
    public static func fetch(url: String, method: String = "GET", headers: [String: String] = [:], body: String? = nil) -> String? {
        var req: [String: Any] = [
            "URL": url,
            "Method": method,
            "Headers": headers,
        ]
        if let body = body {
            req["Body"] = Array(body.utf8)
        }
        guard let jsonData = try? JSONSerialization.data(withJSONObject: req),
              let jsonString = String(data: jsonData, encoding: .utf8) else {
            return nil
        }
        return fetch(jsonString)
    }

    /// Get a value from the KV store.
    public static func kvGet(_ key: String) -> String? {
        var k = key
        return k.withUTF8 { keyBuf in
            guard let keyPtr = keyBuf.baseAddress else { return nil }
            return callHostRead { respPtr, respLenPtr in
                _ff_kv_get(keyPtr, Int32(keyBuf.count), respPtr, respLenPtr)
            }
        }
    }

    /// Set a value in the KV store.
    @discardableResult
    public static func kvSet(_ key: String, _ value: String) -> Bool {
        var k = key
        var v = value
        return k.withUTF8 { keyBuf in
            guard let keyPtr = keyBuf.baseAddress else { return false }
            return v.withUTF8 { valBuf in
                guard let valPtr = valBuf.baseAddress else { return false }
                return _ff_kv_set(keyPtr, Int32(keyBuf.count), valPtr, Int32(valBuf.count)) == 0
            }
        }
    }

    /// Get an environment variable from the host.
    public static func getEnv(_ name: String) -> String? {
        var n = name
        return n.withUTF8 { nameBuf in
            guard let namePtr = nameBuf.baseAddress else { return nil }
            return callHostRead { respPtr, respLenPtr in
                _ff_get_env(namePtr, Int32(nameBuf.count), respPtr, respLenPtr)
            }
        }
    }

    /// Perform AI inference via the FunctionFly AI Gateway.
    ///
    /// - Parameters:
    ///   - model: Model name (e.g. "gpt-4", "claude-3").
    ///   - input: Input text or JSON.
    ///   - params: Optional JSON parameters (temperature, max_tokens, etc.).
    /// - Returns: Model response string, or nil on error.
    public static func aiInfer(model: String, input: String, params: String? = nil) -> String? {
        var m = model
        var inp = input
        return m.withUTF8 { modelBuf in
            guard let modelPtr = modelBuf.baseAddress else { return nil }
            return inp.withUTF8 { inputBuf in
                guard let inputPtr = inputBuf.baseAddress else { return nil }
                if var p = params {
                    return p.withUTF8 { paramsBuf in
                        let paramsPtr = paramsBuf.baseAddress
                        return callHostRead { respPtr, respLenPtr in
                            _ff_ai_infer(modelPtr, Int32(modelBuf.count),
                                        inputPtr, Int32(inputBuf.count),
                                        paramsPtr, Int32(paramsBuf.count),
                                        respPtr, respLenPtr)
                        }
                    }
                } else {
                    return callHostRead { respPtr, respLenPtr in
                        _ff_ai_infer(modelPtr, Int32(modelBuf.count),
                                    inputPtr, Int32(inputBuf.count),
                                    nil, 0,
                                    respPtr, respLenPtr)
                    }
                }
            }
        }
    }
}

// ─── Internal Helpers ────────────────────────────────────────────────────────

/// Call a host function that writes a response to a buffer.
/// Returns the response as a String, or nil on error.
private func callHostRead(_ call: (UnsafeMutablePointer<UInt8>, UnsafeMutablePointer<Int32>) -> Int32) -> String? {
    let maxLen = 1024 * 1024
    let buffer = UnsafeMutablePointer<UInt8>.allocate(capacity: maxLen)
    defer { buffer.deallocate() }

    var respLen: Int32 = 0
    let result = call(buffer, &respLen)

    guard result == 0, respLen > 0, Int(respLen) <= maxLen else {
        return nil
    }

    return String(bytesNoCopy: buffer, length: Int(respLen), encoding: .utf8, freeWhenDone: false)
}
