/// FunctionFly Swift SDK — Host function bindings and runner.
import Foundation

/// Run a FunctionFly function.
public func run(_ function: FunctionFlyFunction) {
    let input = readStdin()
    let context = FunctionFlyContext()

    do {
        let output = try function.handle(input: input, context: context)
        print(output, terminator: "")
    } catch {
        let errorJSON = #"{"error": {"code": "RUNTIME_ERROR", "message": "\#(error.localizedDescription)"}}"#
        fputs("[functionfly] Error: \(error.localizedDescription)\n", stderr)
        print(errorJSON, terminator: "")
        exit(1)
    }
}

/// Read all input from stdin.
private func readStdin() -> String {
    var input = ""
    while let line = readLine() {
        input += line
        input += "\n"
    }
    return input
}
