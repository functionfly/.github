"""Code text builder for FlyEmbed triple-vector embeddings.

Builds AST-aware code representations from function source code and metadata.
Extracts imports, function name, and runtime for code pattern similarity.
"""

import logging
import re
from typing import Any

logger = logging.getLogger(__name__)

# Maximum source code characters to include in embedding text
MAX_SOURCE_CHARS = 2000


class CodeTextBuilder:
    """Builds AST-aware code representation for embedding."""

    def build(self, function_data: dict) -> str:
        """Build code text from function data.

        Produces a structured representation:
            function jwt-verify
            runtime: node18
            imports: jsonwebtoken, crypto
            ---
            // source code content (truncated to MAX_SOURCE_CHARS)

        Args:
            function_data: Dict with source_code, name, runtime, etc.

        Returns:
            Structured code text string
        """
        parts = []
        name = function_data.get("name", "unknown")
        runtime = function_data.get("runtime", "unknown")

        parts.append(f"function {name}")
        parts.append(f"runtime: {runtime}")

        source_code = function_data.get("source_code", "")
        if source_code:
            imports = self._extract_imports(source_code, runtime)
            if imports:
                parts.append(f"imports: {', '.join(imports)}")

            parts.append("---")
            # Truncate to stay within embedding context limits
            parts.append(source_code[:MAX_SOURCE_CHARS])

        return "\n".join(parts)

    def _extract_imports(self, source: str, runtime: str) -> list[str]:
        """Extract import statements from source code.

        Args:
            source: Source code string
            runtime: Runtime identifier (e.g. node18, python3.12)

        Returns:
            List of imported module names (max 10)
        """
        imports: list[str] = []

        if "node" in runtime or "javascript" in runtime or "typescript" in runtime:
            imports = self._extract_js_imports(source)
        elif "python" in runtime:
            imports = self._extract_python_imports(source)
        elif "go" in runtime:
            imports = self._extract_go_imports(source)

        return imports[:10]

    def _extract_js_imports(self, source: str) -> list[str]:
        """Extract imports from JavaScript/TypeScript source."""
        imports: list[str] = []

        # Match require('module') calls
        for match in re.finditer(r"""require\(['"]([^'"]+)['"]\)\s*""", source):
            mod = match.group(1)
            if not mod.startswith("."):
                imports.append(mod)

        # Match import ... from 'module'
        for match in re.finditer(r"""from\s+['"]([^'"]+)['"]\s*""", source):
            mod = match.group(1)
            if not mod.startswith("."):
                imports.append(mod)

        return imports

    def _extract_python_imports(self, source: str) -> list[str]:
        """Extract imports from Python source."""
        imports: list[str] = []

        # Match: from module import ...
        for match in re.finditer(r"^from\s+(\S+)\s+import", source, re.MULTILINE):
            imports.append(match.group(1))

        # Match: import module
        for match in re.finditer(r"^import\s+(\S+)", source, re.MULTILINE):
            imports.append(match.group(1))

        return imports

    def _extract_go_imports(self, source: str) -> list[str]:
        """Extract imports from Go source."""
        imports: list[str] = []

        # Match single import: import "package"
        for match in re.finditer(r"""import\s+"([^"]+)"\s*""", source):
            imports.append(match.group(1))

        # Match import block
        for match in re.finditer(r"""^\s+"([^"]+)"\s*$""", source, re.MULTILINE):
            pkg = match.group(1)
            if "/" in pkg:  # Only external packages
                imports.append(pkg)

        return imports
