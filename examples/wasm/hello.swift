// FunctionFly example: Hello from Swift
// Build: swiftc -target wasm32-unknown-wasi hello.swift -o hello.wasm
// Manifest: {"name":"hello-swift","version":"1.0.0","runtime":"swift","entry":"hello.swift"}
import Foundation

var input = ""
while let line = readLine() {
    input += line + "\n"
}

let response = #"{"message": "Hello from Swift!", "input_length": \#(input.count), "ok": true}"#
print(response, terminator: "")
