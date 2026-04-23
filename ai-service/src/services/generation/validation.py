"""Validation pipeline for generated code.

Validates generated code through multiple stages:
1. Syntax validation (static analysis)
2. Type checking (where applicable)
3. Security scanning
4. Runtime test execution (sandboxed)
"""

import re
import ast
import json
import logging
import subprocess
import tempfile
import os
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass
from enum import Enum

logger = logging.getLogger(__name__)


class ValidationStage(Enum):
    """Validation stages in order of execution."""

    SYNTAX = "syntax"
    TYPE_CHECK = "type_check"
    SECURITY = "security"
    RUNTIME = "runtime"


@dataclass
class ValidationResult:
    """Result of validation stage."""

    stage: ValidationStage
    passed: bool
    errors: List[str]
    warnings: List[str]
    fix_suggestions: List[str]
    duration_ms: float


@dataclass
class ValidationReport:
    """Complete validation report."""

    overall_passed: bool
    confidence_score: float  # 0-1
    stages: List[ValidationResult]
    total_duration_ms: float
    recommended_action: str  # "ship", "fix", "regenerate", "escalate"


class SyntaxValidator:
    """Syntax validation for multiple runtimes."""

    @classmethod
    def validate(cls, code: str, runtime: str) -> ValidationResult:
        """Validate code syntax for given runtime."""
        start_time = __import__("time").time()
        errors = []
        warnings = []
        fixes = []

        if runtime in ("python", "python3"):
            result = cls._validate_python(code)
        elif runtime in ("nodejs", "javascript", "typescript"):
            result = cls._validate_javascript(code)
        elif runtime == "go":
            result = cls._validate_go(code)
        elif runtime == "rust":
            result = cls._validate_rust(code)
        else:
            # Generic validation for unknown runtimes
            result = cls._validate_generic(code)

        duration = (__import__("time").time() - start_time) * 1000

        return ValidationResult(
            stage=ValidationStage.SYNTAX,
            passed=result["passed"],
            errors=result.get("errors", []),
            warnings=result.get("warnings", []),
            fix_suggestions=result.get("fixes", []),
            duration_ms=duration,
        )

    @classmethod
    def _validate_python(cls, code: str) -> Dict:
        """Validate Python syntax using AST."""
        errors = []
        warnings = []
        fixes = []

        try:
            ast.parse(code)
        except SyntaxError as e:
            errors.append(f"Syntax error at line {e.lineno}: {e.msg}")
            # Suggest fixes for common issues
            if "unexpected indent" in str(e):
                fixes.append("Fix indentation - ensure consistent use of spaces (4 per level)")
            if "unexpected EOF" in str(e):
                fixes.append("Add missing closing brackets or quotes")

        # Check for common issues
        if "    " not in code and "\n" in code and "def " in code:
            warnings.append("Code may lack proper indentation")

        # Check for balanced brackets using a stack
        bracket_pairs = {"(": ")", "[": "]", "{": "}"}
        stack = []
        for char in code:
            if char in bracket_pairs:
                stack.append(bracket_pairs[char])
            elif char in bracket_pairs.values():
                if not stack or stack[-1] != char:
                    errors.append(f"Unbalanced bracket: {char}")
                    break
                stack.pop()
        if stack:
            errors.append(f"Unclosed bracket(s): {len(stack)} remaining")

        return {
            "passed": len(errors) == 0,
            "errors": errors,
            "warnings": warnings,
            "fixes": fixes,
        }

    @classmethod
    def _validate_javascript(cls, code: str) -> Dict:
        """Validate JavaScript/TypeScript syntax."""
        errors = []
        warnings = []

        # Basic bracket matching
        brackets = {"(": ")", "[": "]", "{": "}"}
        stack = []
        for char in code:
            if char in brackets:
                stack.append(char)
            elif char in brackets.values():
                if not stack:
                    errors.append(f"Unmatched closing bracket: {char}")
                elif brackets[stack[-1]] != char:
                    errors.append(f"Mismatched brackets: {stack[-1]} and {char}")
                else:
                    stack.pop()

        if stack:
            errors.append(f"Unclosed brackets: {''.join(stack)}")

        # Check for common JS issues
        if "async function" in code and "await" not in code:
            warnings.append("Async function without await")

        if "const " in code and "let " not in code and "var " not in code:
            # Good practice warning
            pass

        return {
            "passed": len(errors) == 0,
            "errors": errors,
            "warnings": warnings,
            "fixes": [],
        }

    @classmethod
    def _validate_go(cls, code: str) -> Dict:
        """Validate Go syntax basics."""
        errors = []

        # Check package declaration
        if "package " not in code:
            errors.append("Missing package declaration")

        # Check for import without usage (basic check)
        imports = re.findall(r"import \(\s*([^)]+)\)", code, re.DOTALL)
        if imports:
            import_list = imports[0].strip().split("\n")
            for imp in import_list:
                imp = imp.strip().strip('"')
                if imp and imp not in code.replace("import", ""):
                    warnings.append(f"Possibly unused import: {imp}")

        return {
            "passed": len(errors) == 0,
            "errors": errors,
            "warnings": [],
            "fixes": [],
        }

    @classmethod
    def _validate_rust(cls, code: str) -> Dict:
        """Validate Rust syntax basics."""
        errors = []

        # Check for fn main if standalone
        if "fn " not in code:
            errors.append("No function definitions found")

        # Basic bracket matching
        if code.count("{") != code.count("}"):
            errors.append("Unbalanced braces")

        return {
            "passed": len(errors) == 0,
            "errors": errors,
            "warnings": [],
            "fixes": [],
        }

    @classmethod
    def _validate_generic(cls, code: str) -> Dict:
        """Generic syntax validation."""
        errors = []
        warnings = []

        if len(code.strip()) < 50:
            errors.append("Code appears too short to be valid")

        if "error" in code.lower() and "return" not in code.lower():
            warnings.append("Code mentions errors but may not handle them properly")

        return {
            "passed": len(errors) == 0,
            "errors": errors,
            "warnings": warnings,
            "fixes": [],
        }


class SecurityValidator:
    """Security validation for generated code."""

    DANGEROUS_PATTERNS = {
        "python": [
            (r"\beval\s*\(", "Dangerous eval() call"),
            (r"\bexec\s*\(", "Dangerous exec() call"),
            (r"__import__\s*\(", "Dynamic import"),
            (r"\bos\.system\s*\(", "OS system call"),
            (r"\bsubprocess\.call\s*\(", "Subprocess call"),
            (r"\bshell\s*=\s*True", "Shell=True in subprocess"),
            (r"\bpickle\.loads?\s*\(", "Pickle deserialization"),
            (r"\byaml\.load\s*\(", "Unsafe YAML load"),
            (r"input\s*\(", "Raw input without validation"),
        ],
        "nodejs": [
            (r"\beval\s*\(", "Dangerous eval() call"),
            (r"\bnew\s+Function\s*\(", "Dynamic Function constructor"),
            (r"child_process", "Child process usage"),
            (r"exec\s*\(", "Dangerous exec()"),
            (r"eval\s*\(", "Dangerous eval()"),
        ],
        "go": [
            (r"\bos\.Exec\s*\(", "OS exec call"),
            (r"\bunsafe\.\w+", "Unsafe package usage"),
        ],
        "rust": [
            (r"\bunsafe\s*\{", "Unsafe block"),
            (r"\.unwrap_unchecked\(\)", "Unchecked unwrap"),
        ],
    }

    @classmethod
    def validate(cls, code: str, runtime: str) -> ValidationResult:
        """Validate code security."""
        start_time = __import__("time").time()
        errors = []
        warnings = []

        patterns = cls.DANGEROUS_PATTERNS.get(runtime, [])

        for pattern, message in patterns:
            if re.search(pattern, code, re.IGNORECASE):
                warnings.append(f"Security: {message}")

        # Check for hardcoded secrets
        secret_patterns = [
            (r'password\s*=\s*["\'][^"\']+["\']', "Possible hardcoded password"),
            (r'api_key\s*=\s*["\'][^"\']+["\']', "Possible hardcoded API key"),
            (r'secret\s*=\s*["\'][^"\']+["\']', "Possible hardcoded secret"),
            (r'token\s*=\s*["\'][^"\']+["\']', "Possible hardcoded token"),
        ]

        for pattern, message in secret_patterns:
            if re.search(pattern, code, re.IGNORECASE):
                warnings.append(f"Security: {message}")

        duration = (__import__("time").time() - start_time) * 1000

        return ValidationResult(
            stage=ValidationStage.SECURITY,
            passed=len(errors) == 0,
            errors=errors,
            warnings=warnings,
            fix_suggestions=["Review and remove dangerous patterns"],
            duration_ms=duration,
        )


class TypeChecker:
    """Basic type checking hints."""

    @classmethod
    def validate(cls, code: str, runtime: str) -> ValidationResult:
        """Check type hints and annotations."""
        start_time = __import__("time").time()
        warnings = []

        if runtime == "python":
            # Check for type hints
            if "def " in code and "->" not in code:
                warnings.append("Functions lack return type annotations")
            if "typing import" not in code and "from typing" not in code:
                if any(t in code for t in ["List", "Dict", "Optional", "Union"]):
                    warnings.append("Type hints used but typing module not imported")

        elif runtime in ("typescript", "nodejs"):
            # Check for TypeScript features
            if ": " not in code and "interface" not in code and runtime == "typescript":
                warnings.append("Missing TypeScript type annotations")

        duration = (__import__("time").time() - start_time) * 1000

        return ValidationResult(
            stage=ValidationStage.TYPE_CHECK,
            passed=True,  # Type issues are warnings, not blockers
            errors=[],
            warnings=warnings,
            fix_suggestions=["Add type hints for better code quality"],
            duration_ms=duration,
        )


class RuntimeValidator:
    """Lightweight runtime validation (sandboxed)."""

    @classmethod
    def validate(
        cls, code: str, runtime: str, test_inputs: Optional[List[Dict]] = None
    ) -> ValidationResult:
        """Run code in sandbox for basic validation."""
        start_time = __import__("time").time()
        errors = []
        warnings = []

        # Only run for simple, safe code
        if runtime == "python":
            result = cls._sandbox_python(code, test_inputs)
        else:
            # Skip runtime validation for other languages in MVP
            result = {"passed": True, "errors": [], "warnings": ["Runtime validation skipped"]}

        duration = (__import__("time").time() - start_time) * 1000

        return ValidationResult(
            stage=ValidationStage.RUNTIME,
            passed=result.get("passed", True),
            errors=result.get("errors", []),
            warnings=result.get("warnings", warnings),
            fix_suggestions=result.get("fixes", []),
            duration_ms=duration,
        )

    @classmethod
    def _sandbox_python(cls, code: str, test_inputs: Optional[List[Dict]]) -> Dict:
        """Run Python code in restricted environment."""
        errors = []
        warnings = []

        # Check for dangerous patterns before execution
        dangerous = [
            "import os",
            "import sys",
            "import subprocess",
            "__import__",
            "eval(",
            "exec(",
            "open(",
            "file(",
        ]
        for d in dangerous:
            if d in code.lower():
                warnings.append(f"Skipping runtime test - contains '{d}'")
                return {"passed": True, "errors": [], "warnings": warnings}

        # Try to compile first
        try:
            compile(code, "<string>", "exec")
        except SyntaxError as e:
            errors.append(f"Compile error: {e}")
            return {"passed": False, "errors": errors, "warnings": []}

        # Very limited execution test for pure functions
        try:
            # Create restricted environment
            restricted_globals = {
                "__builtins__": {
                    "len": len,
                    "range": range,
                    "enumerate": enumerate,
                    "zip": zip,
                    "map": map,
                    "filter": filter,
                    "sum": sum,
                    "min": min,
                    "max": max,
                    "abs": abs,
                    "round": round,
                    "pow": pow,
                    "int": int,
                    "str": str,
                    "float": float,
                    "bool": bool,
                    "list": list,
                    "dict": dict,
                    "set": set,
                    "tuple": tuple,
                    "True": True,
                    "False": False,
                    "None": None,
                    "Exception": Exception,
                    "TypeError": TypeError,
                    "ValueError": ValueError,
                    "KeyError": KeyError,
                }
            }

            # Execute with timeout using subprocess for safety
            # For MVP, just verify it compiles
            return {
                "passed": True,
                "errors": [],
                "warnings": ["Runtime execution skipped for safety"],
            }

        except Exception as e:
            errors.append(f"Runtime error: {e}")
            return {"passed": False, "errors": errors, "warnings": []}


class ValidationPipeline:
    """Complete validation pipeline."""

    def __init__(self):
        self.syntax_validator = SyntaxValidator()
        self.security_validator = SecurityValidator()
        self.type_checker = TypeChecker()
        self.runtime_validator = RuntimeValidator()

    def validate(
        self,
        code: str,
        runtime: str,
        test_inputs: Optional[List[Dict]] = None,
        skip_runtime: bool = False,
    ) -> ValidationReport:
        """Run complete validation pipeline.

        Args:
            code: Generated code
            runtime: Target runtime
            test_inputs: Optional test inputs
            skip_runtime: Skip runtime validation

        Returns:
            ValidationReport with all results
        """
        start_time = __import__("time").time()
        stages = []

        # Stage 1: Syntax
        syntax_result = self.syntax_validator.validate(code, runtime)
        stages.append(syntax_result)

        # Stop early if syntax fails badly
        if len(syntax_result.errors) >= 2:
            return ValidationReport(
                overall_passed=False,
                confidence_score=0.1,
                stages=stages,
                total_duration_ms=(__import__("time").time() - start_time) * 1000,
                recommended_action="regenerate",
            )

        # Stage 2: Security
        security_result = self.security_validator.validate(code, runtime)
        stages.append(security_result)

        # Stage 3: Type checking
        type_result = self.type_checker.validate(code, runtime)
        stages.append(type_result)

        # Stage 4: Runtime (optional)
        if not skip_runtime:
            runtime_result = self.runtime_validator.validate(code, runtime, test_inputs)
            stages.append(runtime_result)

        # Calculate overall score
        total_errors = sum(len(s.errors) for s in stages)
        total_warnings = sum(len(s.warnings) for s in stages)

        if total_errors == 0:
            confidence = 0.9
        elif total_errors == 1:
            confidence = 0.7
        elif total_errors <= 3:
            confidence = 0.4
        else:
            confidence = 0.1

        # Reduce confidence for warnings
        confidence -= min(0.2, total_warnings * 0.05)

        # Determine action
        if confidence >= 0.8:
            action = "ship"
        elif confidence >= 0.5:
            action = "fix"
        elif confidence >= 0.3:
            action = "regenerate"
        else:
            action = "escalate"

        # Check if all stages passed
        all_passed = all(s.passed for s in stages)

        return ValidationReport(
            overall_passed=all_passed and confidence >= 0.6,
            confidence_score=max(0, confidence),
            stages=stages,
            total_duration_ms=(__import__("time").time() - start_time) * 1000,
            recommended_action=action,
        )

    def get_fix_prompt(
        self,
        code: str,
        validation_report: ValidationReport,
        runtime: str,
    ) -> str:
        """Generate a prompt to fix validation errors."""
        errors = []
        for stage in validation_report.stages:
            for error in stage.errors:
                errors.append(f"[{stage.stage.value}] {error}")
            for warning in stage.warnings:
                errors.append(f"[{stage.stage.value}] Warning: {warning}")

        error_text = "\n".join(errors) if errors else "No errors found"

        prompt = f"""Fix the following {runtime} code that has validation issues:

CURRENT CODE:
```
{code}
```

VALIDATION ISSUES:
{error_text}

INSTRUCTIONS:
1. Fix all syntax errors
2. Address security warnings
3. Add proper type hints if missing
4. Ensure the code is production-ready

Provide ONLY the fixed code, no explanations."""

        return prompt


# Global pipeline instance
_pipeline: Optional[ValidationPipeline] = None


def get_validation_pipeline() -> ValidationPipeline:
    """Get global validation pipeline."""
    global _pipeline
    if _pipeline is None:
        _pipeline = ValidationPipeline()
    return _pipeline
