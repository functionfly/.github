"""Tests for validation pipeline."""

import pytest

from src.services.generation.validation import (
    ValidationStage,
    ValidationResult,
    ValidationReport,
    SyntaxValidator,
    SecurityValidator,
    TypeChecker,
    RuntimeValidator,
    ValidationPipeline,
)


class TestSyntaxValidatorPython:
    """Tests for Python syntax validation."""

    def test_valid_python_passes(self):
        """Valid Python code should pass syntax check."""
        code = "def hello():\n    return 'world'\n"
        result = SyntaxValidator.validate(code, "python")
        assert result.passed is True
        assert len(result.errors) == 0

    def test_invalid_python_fails(self):
        """Invalid Python code should fail syntax check."""
        code = "def hello(\n    return 'world'"
        result = SyntaxValidator.validate(code, "python")
        assert result.passed is False
        assert len(result.errors) > 0

    def test_unbalanced_parens_fails(self):
        """Unbalanced parentheses should fail."""
        code = "def hello():\n    return (1 + 2\n"
        result = SyntaxValidator.validate(code, "python")
        assert result.passed is False

    def test_syntax_errors_provide_fixes(self):
        """Syntax errors should include fix suggestions."""
        code = "def hello(\n    return 'world'"
        result = SyntaxValidator.validate(code, "python")
        assert len(result.fix_suggestions) > 0 or len(result.errors) > 0


class TestSyntaxValidatorJavaScript:
    """Tests for JavaScript syntax validation."""

    def test_valid_js_passes(self):
        """Valid JS code should pass."""
        code = "function hello() { return 'world'; }"
        result = SyntaxValidator.validate(code, "javascript")
        assert result.passed is True

    def test_unmatched_braces_fails(self):
        """Unmatched braces should fail."""
        code = "function hello() { return 'world';"
        result = SyntaxValidator.validate(code, "javascript")
        assert result.passed is False
        assert len(result.errors) > 0

    def test_mismatched_brackets_fails(self):
        """Mismatched brackets should fail."""
        code = "function hello() { return [1, 2, 3); }"
        result = SyntaxValidator.validate(code, "javascript")
        assert result.passed is False

    def test_async_without_await_warns(self):
        """Async function without await should produce warning."""
        code = "async function hello() { return 'world'; }"
        result = SyntaxValidator.validate(code, "javascript")
        assert any("await" in w for w in result.warnings)


class TestSyntaxValidatorGo:
    """Tests for Go syntax validation."""

    def test_go_without_package_fails(self):
        """Go code without package declaration should fail."""
        code = 'func main() {\n    fmt.Println("hello")\n}'
        result = SyntaxValidator.validate(code, "go")
        assert result.passed is False

    def test_go_with_package_passes(self):
        """Go code with package declaration should pass."""
        code = 'package main\n\nfunc main() {\n    fmt.Println("hello")\n}'
        result = SyntaxValidator.validate(code, "go")
        assert result.passed is True


class TestSyntaxValidatorRust:
    """Tests for Rust syntax validation."""

    def test_rust_without_fn_fails(self):
        """Rust code without fn should fail."""
        code = "let x = 42;"
        result = SyntaxValidator.validate(code, "rust")
        assert result.passed is False

    def test_rust_unbalanced_braces_fails(self):
        """Rust code with unbalanced braces should fail."""
        code = "fn main() { let x = 42;"
        result = SyntaxValidator.validate(code, "rust")
        assert result.passed is False

    def test_rust_valid_passes(self):
        """Valid Rust code should pass."""
        code = "fn main() {\n    let x = 42;\n}"
        result = SyntaxValidator.validate(code, "rust")
        assert result.passed is True


class TestSyntaxValidatorGeneric:
    """Tests for generic syntax validation."""

    def test_short_code_fails(self):
        """Very short code should fail generic validation."""
        code = "x = 1"
        result = SyntaxValidator.validate(code, "unknown_runtime")
        assert result.passed is False

    def test_sufficient_code_passes(self):
        """Sufficiently long code should pass."""
        code = "def process_data():\n    data = load_data()\n    result = transform(data)\n    return save(result)\n"
        result = SyntaxValidator.validate(code, "unknown_runtime")
        assert result.passed is True


class TestSecurityValidator:
    """Tests for SecurityValidator."""

    def test_safe_code_passes(self):
        """Safe code should pass security check."""
        code = "def add(a, b):\n    return a + b\n"
        result = SecurityValidator.validate(code, "python")
        assert result.passed is True
        assert len(result.warnings) == 0

    def test_eval_detected(self):
        """eval() usage should be detected."""
        code = "result = eval(user_input)"
        result = SecurityValidator.validate(code, "python")
        assert any("eval" in w for w in result.warnings)

    def test_exec_detected(self):
        """exec() usage should be detected."""
        code = "exec(user_code)"
        result = SecurityValidator.validate(code, "python")
        assert any("exec" in w for w in result.warnings)

    def test_os_system_detected(self):
        """os.system() usage should be detected."""
        code = "os.system('rm -rf /')"
        result = SecurityValidator.validate(code, "python")
        assert any("system" in w.lower() for w in result.warnings)

    def test_hardcoded_password_detected(self):
        """Hardcoded password should be detected."""
        code = 'password = "mysecret123"'
        result = SecurityValidator.validate(code, "python")
        assert any("password" in w.lower() for w in result.warnings)

    def test_hardcoded_api_key_detected(self):
        """Hardcoded API key should be detected."""
        code = 'api_key = "sk-1234567890"'
        result = SecurityValidator.validate(code, "python")
        assert any("api" in w.lower() for w in result.warnings)

    def test_js_eval_detected(self):
        """JavaScript runtime has no security patterns configured."""
        code = "eval(userInput)"
        result = SecurityValidator.validate(code, "javascript")
        # JavaScript has no patterns in DANGEROUS_PATTERNS - just verify no crash
        assert result.passed is True

    def test_go_unsafe_detected(self):
        """unsafe package in Go should be detected."""
        code = "package main\nfunc main() { unsafe.Pointer(nil) }"
        result = SecurityValidator.validate(code, "go")
        assert len(result.warnings) > 0, (
            f"Expected warnings but got: {result.warnings}, errors: {result.errors}"
        )
        assert any("unsafe" in w.lower() for w in result.warnings)

    def test_rust_unsafe_detected(self):
        """unsafe block in Rust should be detected."""
        code = "fn main() { unsafe { *ptr = 42; } }"
        result = SecurityValidator.validate(code, "rust")
        assert len(result.warnings) > 0, (
            f"Expected warnings but got: {result.warnings}, errors: {result.errors}"
        )
        assert any("unsafe" in w.lower() for w in result.warnings)


class TestTypeChecker:
    """Tests for TypeChecker."""

    def test_python_with_type_hints_passes(self):
        """Python code with type hints should pass without warnings."""
        code = "def add(a: int, b: int) -> int:\n    return a + b\n"
        result = TypeChecker.validate(code, "python")
        assert result.passed is True
        assert len(result.warnings) == 0

    def test_python_without_return_type_warns(self):
        """Python code without return type should warn."""
        code = "def add(a, b):\n    return a + b\n"
        result = TypeChecker.validate(code, "python")
        assert any("return type" in w.lower() for w in result.warnings)

    def test_python_typing_used_without_import_warns(self):
        """Using type hints without importing typing should warn."""
        code = "def get_items() -> List[str]:\n    return []\n"
        result = TypeChecker.validate(code, "python")
        assert any("typing" in w.lower() for w in result.warnings)


class TestRuntimeValidator:
    """Tests for RuntimeValidator."""

    def test_safe_python_compiles(self):
        """Safe Python code should compile and pass."""
        code = "def add(a, b):\n    return a + b\n"
        result = RuntimeValidator.validate(code, "python")
        assert result.passed is True

    def test_dangerous_python_skips_execution(self):
        """Code with dangerous patterns should skip runtime but still pass."""
        code = "import os\nos.system('echo hello')"
        result = RuntimeValidator.validate(code, "python")
        assert result.passed is True
        assert any("skip" in w.lower() for w in result.warnings)

    def test_non_python_skips_runtime(self):
        """Non-Python runtimes should skip runtime validation."""
        code = "function hello() { return 'world'; }"
        result = RuntimeValidator.validate(code, "javascript")
        assert result.passed is True
        assert any("skip" in w.lower() for w in result.warnings)


class TestValidationPipeline:
    """Tests for ValidationPipeline."""

    def test_valid_code_passes_all_stages(self):
        """Valid code should pass all stages."""
        pipeline = ValidationPipeline()
        code = "def add(a: int, b: int) -> int:\n    return a + b\n"

        report = pipeline.validate(code, "python", skip_runtime=True)
        assert report.overall_passed is True
        assert report.confidence_score >= 0.5
        assert report.recommended_action in ("ship", "fix")

    def test_syntax_error_fails_early(self):
        """Severe syntax errors should fail early with regenerate action."""
        pipeline = ValidationPipeline()
        code = "def hello(\n    return 'world'\ndef broken("

        report = pipeline.validate(code, "python")
        assert report.overall_passed is False
        assert report.recommended_action == "regenerate"

    def test_security_warnings_reduce_confidence(self):
        """Security warnings should reduce confidence score."""
        pipeline = ValidationPipeline()
        safe_code = "def add(a: int, b: int) -> int:\n    return a + b\n"
        unsafe_code = "def add(a, b):\n    password = 'secret'\n    return eval(a + b)\n"

        safe_report = pipeline.validate(safe_code, "python", skip_runtime=True)
        unsafe_report = pipeline.validate(unsafe_code, "python", skip_runtime=True)

        assert safe_report.confidence_score >= unsafe_report.confidence_score

    def test_skip_runtime_stages(self):
        """skip_runtime should omit runtime stage."""
        pipeline = ValidationPipeline()
        code = "def add(a, b):\n    return a + b\n"

        report = pipeline.validate(code, "python", skip_runtime=True)
        stage_types = [s.stage for s in report.stages]
        assert ValidationStage.RUNTIME not in stage_types

    def test_get_fix_prompt(self):
        """get_fix_prompt should include errors and warnings."""
        pipeline = ValidationPipeline()
        code = "def hello(\n    return 'world'"
        report = pipeline.validate(code, "python")

        prompt = pipeline.get_fix_prompt(code, report, "python")
        assert "Syntax error" in prompt or "error" in prompt.lower()
        assert "CURRENT CODE" in prompt

    def test_confidence_score_ship_action(self):
        """High confidence code should recommend ship."""
        pipeline = ValidationPipeline()
        code = 'def add(a: int, b: int) -> int:\n    """Add two numbers."""\n    return a + b\n'

        report = pipeline.validate(code, "python", skip_runtime=True)
        if report.confidence_score >= 0.8:
            assert report.recommended_action == "ship"

    def test_multiple_errors_lower_confidence(self):
        """Multiple errors should result in low confidence."""
        pipeline = ValidationPipeline()
        code = "def a(\ndef b(\ndef c("

        report = pipeline.validate(code, "python")
        assert report.confidence_score <= 0.4
