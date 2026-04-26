/// FunctionFly Swift SDK — Execution context.
import Foundation

/// Execution context providing access to host functions.
public struct FunctionFlyContext {
    public let functionName: String
    public let version: String

    public init() {
        self.functionName = ProcessInfo.processInfo.environment["FUNCTIONFLY_FUNCTION_NAME"] ?? ""
        self.version = ProcessInfo.processInfo.environment["FUNCTIONFLY_VERSION"] ?? ""
    }

    /// Log a message.
    public func log(_ message: String) {
        fputs("[functionfly] \(message)\n", stderr)
    }

    /// Get an environment variable.
    public func getEnv(_ key: String) -> String? {
        ProcessInfo.processInfo.environment[key]
    }

    /// Return an error JSON response.
    public func error(code: String, message: String) -> String {
        return #"{"error": {"code": "\#(code)", "message": "\#(message)"}}"#
    }

    /// Return a success JSON response.
    public func json(_ data: String) -> String {
        return #"{"ok": true, "data": \#(data)}"#
    }
}
