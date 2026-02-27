"""
Build system for FlyPy functions.

Handles serialization of function metadata and integration with the Go compiler.
"""

import json
import os
import tempfile
from pathlib import Path
from typing import Dict, Any, Optional, List
import ast

from .decorators import get_registered_functions, get_function_definition
from .types import FunctionDefinition, BuildResult
from .optimizer import optimize_bundle, analyze_bundle_size, MinimalRuntimeBuilder
from .cold_start_optimizer import optimize_function_cold_start
from .build_optimizer import optimize_build_process


class FlyPyBuilder:
    """Builder for FlyPy functions."""

    def __init__(
        self,
        go_binary_path: Optional[str] = None,
        optimize_bundle: bool = True,
        optimization_level: str = "balanced",
        optimize_cold_start: bool = True,
        enable_parallel: bool = True,
        enable_incremental: bool = True,
        max_parallel_workers: Optional[int] = None
    ):
        """
        Initialize the builder.

        Args:
            go_binary_path: Path to the FlyPy Go binary
            optimize_bundle: Whether to optimize bundle size
            optimization_level: Optimization level ("minimal", "balanced", "aggressive")
            optimize_cold_start: Whether to optimize cold start performance
            enable_parallel: Whether to enable parallel building
            enable_incremental: Whether to enable incremental builds
            max_parallel_workers: Maximum number of parallel workers
        """
        self.go_binary_path = go_binary_path or self._find_go_binary()
        self.optimize_bundle = optimize_bundle
        self.optimization_level = optimization_level
        self.optimize_cold_start = optimize_cold_start
        self.enable_parallel = enable_parallel
        self.enable_incremental = enable_incremental
        self.max_parallel_workers = max_parallel_workers
        self.runtime_builder = MinimalRuntimeBuilder()

    def build_function(
        self,
        func_name: str,
        output_dir: str = "./dist",
        mode: str = "deterministic",
        verbose: bool = False,
        optimize: Optional[bool] = None,
        optimization_level: Optional[str] = None,
        optimize_cold_start: Optional[bool] = None
    ) -> BuildResult:
        """
        Build a single FlyPy function.

        Args:
            func_name: Name of the function to build
            output_dir: Output directory for artifacts
            mode: Execution mode ("deterministic" or "compatible")
            verbose: Enable verbose output
            optimize: Override bundle optimization setting
            optimization_level: Override optimization level
            optimize_cold_start: Override cold start optimization setting

        Returns:
            BuildResult with build information
        """
        func_def = get_function_definition(func_name)
        if not func_def:
            return BuildResult(
                success=False,
                function_name=func_name,
                output_dir=output_dir,
                errors=[f"Function '{func_name}' not found in registry"]
            )

        # Use provided optimization settings or instance defaults
        should_optimize = optimize if optimize is not None else self.optimize_bundle
        opt_level = optimization_level or self.optimization_level
        should_optimize_cold_start = optimize_cold_start if optimize_cold_start is not None else self.optimize_cold_start

        return self._build_function_definition(
            func_def, output_dir, mode, verbose, should_optimize, opt_level, should_optimize_cold_start
        )

    def build_all_functions(
        self,
        output_dir: str = "./dist",
        mode: str = "deterministic",
        verbose: bool = False,
        parallel: Optional[bool] = None,
        incremental: Optional[bool] = None
    ) -> List[BuildResult]:
        """
        Build all registered FlyPy functions.

        Args:
            output_dir: Output directory for artifacts
            mode: Execution mode ("deterministic" or "compatible")
            verbose: Enable verbose output
            parallel: Override parallel building setting
            incremental: Override incremental build setting

        Returns:
            List of BuildResult objects
        """
        functions = get_registered_functions()
        if not functions:
            return [BuildResult(
                success=False,
                function_name="",
                output_dir=output_dir,
                errors=["No FlyPy functions found in registry"]
            )]

        # Use provided settings or instance defaults
        use_parallel = parallel if parallel is not None else self.enable_parallel
        use_incremental = incremental if incremental is not None else self.enable_incremental

        # Check if we should use optimized building
        if use_parallel or use_incremental:
            # Convert function definitions to dict format for optimizer
            func_defs = []
            for func_name, func_def in functions.items():
                func_dict = {
                    "metadata": func_def.metadata.model_dump(),
                    "source": {
                        "code": func_def.source_code,
                        "file": func_def.metadata.source_file,
                        "dependencies": func_def.dependencies,
                        "imports": func_def.imports,
                    },
                    "ast": None  # Will be generated if needed
                }
                func_defs.append(func_dict)

            # Build configuration
            build_config = {
                "mode": mode,
                "optimize_bundle": self.optimize_bundle,
                "optimization_level": self.optimization_level,
                "optimize_cold_start": self.optimize_cold_start,
            }

            # Use build optimizer
            optimization_results = optimize_build_process(
                func_defs,
                self.go_binary_path,
                output_dir,
                build_config,
                enable_parallel=use_parallel,
                enable_incremental=use_incremental,
                verbose=verbose
            )

            # Convert results back to BuildResult objects
            results = []
            for result in optimization_results["results"]:
                build_result = BuildResult(
                    success=result["success"],
                    function_name=result["function_name"],
                    output_dir=result["output_dir"],
                    warnings=result.get("warnings", []),
                    errors=result.get("errors", []),
                    wasm_file=result.get("wasm_file"),
                    manifest_file=result.get("manifest_file"),
                    build_time_ms=result.get("build_time_ms"),
                    wasm_size_bytes=result.get("wasm_size_bytes"),
                    optimization_stats=result.get("optimization_stats", {}),
                    bundle_analysis=result.get("bundle_analysis", {}),
                    cold_start_stats=result.get("cold_start_stats", {})
                )
                results.append(build_result)

            return results

        else:
            # Fall back to sequential building
            results = []
            for func_name, func_def in functions.items():
                result = self._build_function_definition(
                    func_def, output_dir, mode, verbose,
                    self.optimize_bundle, self.optimization_level, self.optimize_cold_start
                )
                results.append(result)

            return results

    def _build_function_definition(
        self,
        func_def: FunctionDefinition,
        output_dir: str,
        mode: str,
        verbose: bool,
        optimize_bundle: bool = True,
        optimization_level: str = "balanced",
        optimize_cold_start: bool = True
    ) -> BuildResult:
        """Build a function definition."""
        import subprocess
        import time

        func_name = func_def.metadata.name
        start_time = time.time()

        try:
            # Create output directory
            output_path = Path(output_dir) / func_name
            output_path.mkdir(parents=True, exist_ok=True)

            # Create metadata file for Go compiler
            metadata_file = output_path / "function_metadata.json"
            with open(metadata_file, 'w') as f:
                json.dump(self._serialize_function_definition(func_def), f, indent=2)

            # Optimize source code if requested
            source_code = func_def.source_code
            optimization_stats = {}

            if optimize_bundle:
                if verbose:
                    print(f"Optimizing bundle with level: {optimization_level}")

                optimized_code, opt_stats = optimize_bundle(
                    func_def.source_code,
                    func_def.metadata.dict(),
                    func_def.dependencies
                )
                source_code = optimized_code
                optimization_stats = opt_stats

                if verbose:
                    print(f"Bundle optimization: {opt_stats}")

            # Create source file
            source_file = output_path / "function.py"
            with open(source_file, 'w') as f:
                f.write(source_code)

            # Check if Go binary is available
            if not self.go_binary_path or not Path(self.go_binary_path).exists():
                return BuildResult(
                    success=False,
                    function_name=func_name,
                    output_dir=str(output_path),
                    errors=["FlyPy Go binary not found"]
                )

            # Run Go compiler
            cmd = [
                self.go_binary_path,
                "--input", str(source_file),
                "--metadata", str(metadata_file),
                "--output", str(output_path),
                "--mode", mode,
            ]

            # Filter out None values
            cmd = [arg for arg in cmd if arg is not None]

            if verbose:
                cmd.append("--verbose")
                print(f"Running: {' '.join(cmd)}")

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=os.getcwd()
            )

            build_time = int((time.time() - start_time) * 1000)

            if result.returncode == 0:
                # Check for artifacts
                wasm_file = output_path / "state_transition.wasm"
                manifest_file = output_path / "manifest.json"

                wasm_size = None
                if wasm_file.exists():
                    wasm_size = wasm_file.stat().st_size

                warnings = []
                if result.stderr:
                    warnings.extend(result.stderr.strip().split('\n'))

                # Analyze final bundle size
                bundle_analysis = {}
                if wasm_file.exists():
                    with open(wasm_file, 'rb') as f:
                        bundle_content = f.read()
                    bundle_analysis = analyze_bundle_size(bundle_content.decode('latin-1', errors='ignore'), func_def.metadata.dict())

                # Apply cold start optimizations if configured
                cold_start_stats = {}
                if optimize_cold_start:
                    # Check if function has cold start config
                    cold_start_config = getattr(func_def.metadata, '_cold_start_config', {})
                    if cold_start_config.get("optimize_cold_start", True):
                        cold_start_stats = optimize_function_cold_start(
                            func_name,
                            func_def.source_code,
                            func_def.dependencies,
                            cold_start_config.get("warmup_data", [])
                        )

                return BuildResult(
                    success=True,
                    function_name=func_name,
                    output_dir=str(output_path),
                    warnings=warnings,
                    wasm_file=str(wasm_file) if wasm_file.exists() else None,
                    manifest_file=str(manifest_file) if manifest_file.exists() else None,
                    build_time_ms=build_time,
                    wasm_size_bytes=wasm_size,
                    optimization_stats=optimization_stats,
                    bundle_analysis=bundle_analysis,
                    cold_start_stats=cold_start_stats,
                )
            else:
                errors = []
                if result.stderr:
                    errors.extend(result.stderr.strip().split('\n'))
                if not errors and result.stdout:
                    errors.extend(result.stdout.strip().split('\n'))

                return BuildResult(
                    success=False,
                    function_name=func_name,
                    output_dir=str(output_path),
                    errors=errors,
                    build_time_ms=build_time,
                )

        except Exception as e:
            build_time = int((time.time() - start_time) * 1000)
            return BuildResult(
                success=False,
                function_name=func_def.metadata.name,
                output_dir=output_dir,
                errors=[f"Build failed: {str(e)}"],
                build_time_ms=build_time,
            )

    def _serialize_function_definition(self, func_def: FunctionDefinition) -> Dict[str, Any]:
        """Serialize a function definition for the Go compiler."""
        # Parse the AST to extract more detailed information
        try:
            tree = ast.parse(func_def.source_code)
            ast_info = self._analyze_ast(tree, func_def.metadata.name)
        except SyntaxError:
            ast_info = {}

        return {
            "metadata": {
                "name": func_def.metadata.name,
                "version": func_def.metadata.version,
                "description": func_def.metadata.description,
                "deterministic": func_def.metadata.deterministic,
                "idempotent": func_def.metadata.idempotent,
                "pure": func_def.metadata.pure,
                "cache_ttl": func_def.metadata.cache_ttl,
                "capabilities": func_def.metadata.capabilities,
                "max_execution_time": func_def.metadata.max_execution_time,
                "input_schema": func_def.metadata.input_schema,
                "output_schema": func_def.metadata.output_schema,
                "execution_mode": func_def.metadata.execution_mode,
                "source_hash": func_def.metadata.source_hash,
            },
            "source": {
                "code": func_def.source_code,
                "file": func_def.metadata.source_file,
                "dependencies": func_def.dependencies,
                "imports": func_def.imports,
            },
            "ast": ast_info,
        }

    def _analyze_ast(self, tree: ast.AST, func_name: str) -> Dict[str, Any]:
        """Analyze AST to extract function information."""
        analyzer = ASTAnalyzer(func_name)
        analyzer.visit(tree)
        return analyzer.get_info()

    def _find_go_binary(self) -> Optional[str]:
        """Find the FlyPy Go binary."""
        import shutil

        # Check PATH
        go_binary = os.path.expanduser("~/bin/flypy-go")
        if go_binary:
            return go_binary

        # Check relative to this file
        script_dir = Path(__file__).parent.parent.parent.parent
        candidates = [
            script_dir / "cmd" / "flypy" / "flypy",
            script_dir / "bin" / "flypy",
            script_dir / "flypy",
        ]

        for candidate in candidates:
            if candidate.exists() and candidate.is_file():
                return str(candidate)

        return None


class ASTAnalyzer(ast.NodeVisitor):
    """AST analyzer for FlyPy functions."""

    def __init__(self, func_name: str):
        self.func_name = func_name
        self.functions = {}
        self.current_function = None
        self.globals = set()
        self.imports = []

    def visit_FunctionDef(self, node):
        if node.name == self.func_name:
            self.current_function = {
                "name": node.name,
                "args": [arg.arg for arg in node.args.args],
                "line_start": node.lineno,
                "line_end": getattr(node, 'end_lineno', node.lineno),
                "decorators": [self._get_decorator_name(d) for d in node.decorator_list],
            }
            self.functions[node.name] = self.current_function
        self.generic_visit(node)

    def visit_Global(self, node):
        if self.current_function:
            self.globals.update(node.names)
        self.generic_visit(node)

    def visit_Import(self, node):
        for alias in node.names:
            self.imports.append({
                "type": "import",
                "module": alias.name,
                "alias": alias.asname,
            })
        self.generic_visit(node)

    def visit_ImportFrom(self, node):
        module = node.module or ""
        for alias in node.names:
            self.imports.append({
                "type": "from_import",
                "module": module,
                "name": alias.name,
                "alias": alias.asname,
            })
        self.generic_visit(node)

    def _get_decorator_name(self, decorator):
        """Get decorator name from AST node."""
        if isinstance(decorator, ast.Name):
            return decorator.id
        elif isinstance(decorator, ast.Attribute):
            return f"{decorator.attr}"
        elif isinstance(decorator, ast.Call):
            return self._get_decorator_name(decorator.func)
        return str(decorator)

    def get_info(self) -> Dict[str, Any]:
        """Get analysis information."""
        return {
            "functions": self.functions,
            "globals": list(self.globals),
            "imports": self.imports,
        }


def build_function(
    func_name: str,
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False,
    go_binary: Optional[str] = None
) -> BuildResult:
    """
    Convenience function to build a single FlyPy function.

    Args:
        func_name: Name of the function to build
        output_dir: Output directory for artifacts
        mode: Execution mode ("deterministic" or "compatible")
        verbose: Enable verbose output
        go_binary: Path to FlyPy Go binary

    Returns:
        BuildResult with build information
    """
    builder = FlyPyBuilder(go_binary)
    return builder.build_function(func_name, output_dir, mode, verbose)


def build_all(
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False,
    go_binary: Optional[str] = None,
    optimize_bundle: bool = True,
    optimization_level: str = "balanced",
    optimize_cold_start: bool = True,
    enable_parallel: bool = True,
    enable_incremental: bool = True,
    max_parallel_workers: Optional[int] = None
) -> List[BuildResult]:
    """
    Convenience function to build all registered FlyPy functions.

    Args:
        output_dir: Output directory for artifacts
        mode: Execution mode ("deterministic" or "compatible")
        verbose: Enable verbose output
        go_binary: Path to FlyPy Go binary
        optimize_bundle: Whether to optimize bundle size
        optimization_level: Bundle optimization level
        optimize_cold_start: Whether to optimize cold start
        enable_parallel: Whether to enable parallel building
        enable_incremental: Whether to enable incremental builds
        max_parallel_workers: Maximum parallel workers

    Returns:
        List of BuildResult objects
    """
    builder = FlyPyBuilder(
        go_binary_path=go_binary,
        optimize_bundle=optimize_bundle,
        optimization_level=optimization_level,
        optimize_cold_start=optimize_cold_start,
        enable_parallel=enable_parallel,
        enable_incremental=enable_incremental,
        max_parallel_workers=max_parallel_workers
    )
    return builder.build_all_functions(output_dir, mode, verbose)