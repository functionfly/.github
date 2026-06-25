import XCTest
@testable import FunctionFly

final class FunctionFlyTests: XCTestCase {

    // ─── FunctionFlyContext Tests ──────────────────────────────────────────────

    func testContextErrorJSON() {
        let ctx = FunctionFlyContext()
        let json = ctx.error(code: "TEST_ERROR", message: "something went wrong")

        XCTAssertTrue(json.contains("\"error\""))
        XCTAssertTrue(json.contains("TEST_ERROR"))
        XCTAssertTrue(json.contains("something went wrong"))
    }

    func testContextSuccessJSON() {
        let ctx = FunctionFlyContext()
        let json = ctx.json(#"{"name": "test"}"#)

        XCTAssertTrue(json.contains("\"ok\": true"))
        XCTAssertTrue(json.contains("\"data\""))
        XCTAssertTrue(json.contains("\"name\": \"test\""))
    }

    func testContextFunctionNameFromEnv() {
        // functionName reads from FUNCTIONFLY_FUNCTION_NAME env var
        // In test environment it will be empty unless set
        let ctx = FunctionFlyContext()
        // Just verify it doesn't crash and returns a string
        XCTAssertNotNil(ctx.functionName)
    }

    func testContextVersionFromEnv() {
        let ctx = FunctionFlyContext()
        XCTAssertNotNil(ctx.version)
    }

    func testContextGetEnv() {
        let ctx = FunctionFlyContext()
        // PATH should exist in any test environment
        let path = ctx.getEnv("PATH")
        XCTAssertNotNil(path)
        XCTAssertFalse(path?.isEmpty ?? true)
    }

    func testContextGetEnv_Nonexistent() {
        let ctx = FunctionFlyContext()
        let value = ctx.getEnv("DEFINITELY_NOT_A_REAL_ENV_VAR_12345")
        XCTAssertNil(value)
    }

    // ─── FunctionFlyFunction Protocol Tests ───────────────────────────────────

    func testFunctionProtocol_Conformance() {
        let fn = EchoFunction()
        let ctx = FunctionFlyContext()

        let result = try! fn.handle(input: #"{"hello": "world"}"#, context: ctx)
        XCTAssertTrue(result.contains("hello"))
        XCTAssertTrue(result.contains("world"))
    }

    func testFunctionProtocol_ThrowingFunction() {
        let fn = ThrowingFunction()
        let ctx = FunctionFlyContext()

        XCTAssertThrowsError(try fn.handle(input: "anything", context: ctx)) { error in
            XCTAssertTrue(error is FunctionError)
        }
    }

    func testFunctionProtocol_EmptyInput() {
        let fn = EchoFunction()
        let ctx = FunctionFlyContext()

        let result = try! fn.handle(input: "", context: ctx)
        XCTAssertNotNil(result)
    }

    func testFunctionProtocol_LargeInput() {
        let fn = EchoFunction()
        let ctx = FunctionFlyContext()

        let largeInput = String(repeating: "x", count: 100_000)
        let result = try! fn.handle(input: largeInput, context: ctx)
        XCTAssertTrue(result.contains("x"))
    }

    // ─── Error JSON Format Tests ──────────────────────────────────────────────

    func testErrorJSON_Format() {
        let ctx = FunctionFlyContext()
        let json = ctx.error(code: "VALIDATION_ERROR", message: "invalid input")

        // Should be valid JSON structure
        XCTAssertTrue(json.hasPrefix("{"))
        XCTAssertTrue(json.hasSuffix("}"))
        XCTAssertTrue(json.contains("\"code\""))
        XCTAssertTrue(json.contains("\"message\""))
    }

    func testErrorJSON_SpecialCharacters() {
        let ctx = FunctionFlyContext()
        let json = ctx.error(code: "ERR", message: "line1\nline2")

        // The SDK uses raw string interpolation — newlines are passed through as-is
        XCTAssertTrue(json.contains("line1"))
        XCTAssertTrue(json.contains("line2"))
        XCTAssertTrue(json.contains("ERR"))
    }

    // ─── Success JSON Format Tests ────────────────────────────────────────────

    func testSuccessJSON_Format() {
        let ctx = FunctionFlyContext()
        let json = ctx.json(#"{"key": "value"}"#)

        XCTAssertTrue(json.hasPrefix("{"))
        XCTAssertTrue(json.hasSuffix("}"))
        XCTAssertTrue(json.contains("\"ok\": true"))
    }

    func testSuccessJSON_EmptyData() {
        let ctx = FunctionFlyContext()
        let json = ctx.json("{}")

        XCTAssertTrue(json.contains("\"ok\": true"))
        XCTAssertTrue(json.contains("{}"))
    }

    // ─── Log Function Tests ───────────────────────────────────────────────────

    func testContextLog_DoesNotCrash() {
        let ctx = FunctionFlyContext()
        // log() writes to stderr — just verify it doesn't crash
        ctx.log("test message")
        ctx.log("")
        ctx.log("multi\nline\nmessage")
    }
}

// ─── Test Helpers ─────────────────────────────────────────────────────────────

/// A simple function that echoes input back.
struct EchoFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        return context.json(#"{"echo": "\#(input)"}"#)
    }
}

/// An error type for testing.
enum FunctionError: Error {
    case intentional
}

/// A function that always throws.
struct ThrowingFunction: FunctionFlyFunction {
    func handle(input: String, context: FunctionFlyContext) throws -> String {
        throw FunctionError.intentional
    }
}
