/// Example FunctionFly function in Swift.
/// Build: swiftc -target wasm32-unknown-wasi main.swift -o hello.wasm
import FunctionFly

struct HelloFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        context.log("Hello from Swift function!")
        return #"{"message": "Hello from Swift!", "ok": true}"#
    }
}

FunctionFly.run(HelloFunction())
