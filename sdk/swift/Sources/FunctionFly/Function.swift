/// FunctionFly Swift SDK — Function protocol.
///
/// Implement this protocol to create a FunctionFly function:
///
///     struct MyFunction: FunctionFlyFunction {
///         func handle(input: String, context: FunctionFlyContext) throws -> String {
///             return #"{"message": "Hello from Swift!"}"#
///         }
///     }
///
///     FunctionFly.run(MyFunction())
import Foundation

/// Protocol that all FunctionFly functions must implement.
public protocol FunctionFlyFunction {
    /// Process the input and return the output.
    func handle(input: String, context: FunctionFlyContext) throws -> String
}
