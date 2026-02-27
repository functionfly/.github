"""
Decorators for FlyPy functions.

Provides the @function decorator and schema decorators for defining
deterministic Python functions that can be compiled to WebAssembly.
"""

import functools
import inspect
import hashlib
import ast
import sys
from typing import Dict, List, Any, Optional, Callable, Union
from pathlib import Path

from .types import FunctionMetadata, FunctionDefinition, ExecutionMode
from .schema import Schema
from .cold_start_optimizer import optimize_function_cold_start
from .performance_monitor import performance_monitor


# Global registry of FlyPy functions
_FUNCTION_REGISTRY: Dict[str, FunctionDefinition] = {}


def function(
    name: Optional[str] = None,
    version: str = "1.0.0",
    description: Optional[str] = None,
    deterministic: bool = True,
    idempotent: bool = False,
    pure: bool = False,
    cache_ttl: Optional[int] = None,
    capabilities: Optional[List[str]] = None,
    max_execution_time: Optional[int] = None,
    execution_mode: ExecutionMode = ExecutionMode.DETERMINISTIC,
    optimize_cold_start: bool = True,
    warmup_data: Optional[List[Dict[str, Any]]] = None,
    enable_performance_monitoring: bool = False,
) -> Callable:
    """
    Decorator to mark a function as a FlyPy function.

    Args:
        name: Function name (defaults to function.__name__)
        version: Semantic version string
        description: Function description
        deterministic: Whether the function is deterministic
        idempotent: Whether the function is idempotent (same input always produces same output)
        pure: Whether the function is pure (no side effects)
        cache_ttl: Cache TTL in seconds
        capabilities: List of required capabilities (e.g., ["network", "filesystem"])
        max_execution_time: Maximum execution time in milliseconds
        execution_mode: Execution mode (deterministic or compatible)
        optimize_cold_start: Whether to apply cold start optimizations
        warmup_data: Test data for function warm-up to reduce cold starts

    Returns:
        Decorated function

    Example:
        @flypy.function(
            name="calculate-total",
            deterministic=True,
            idempotent=True,
            cache_ttl=3600,
            capabilities=["math"]
        )
        def handler(event):
            return {"total": sum(event["items"])}
    """

    def decorator(func: Callable) -> Callable:
        # Determine function name
        func_name = name if name is not None else func.__name__

        # Get source code and hash
        source_code = _get_function_source(func)
        source_hash = hashlib.sha256(source_code.encode()).hexdigest()

        # Get source file path
        source_file = None
        try:
            source_file = inspect.getfile(func)
        except (OSError, TypeError):
            pass

        # Create metadata
        metadata = FunctionMetadata(
            name=func_name,
            version=version,
            description=description or func.__doc__,
            deterministic=deterministic,
            idempotent=idempotent,
            pure=pure,
            cache_ttl=cache_ttl,
            capabilities=capabilities or [],
            max_execution_time=max_execution_time,
            execution_mode=execution_mode,
            source_file=source_file,
            source_hash=source_hash,
        )

        # Store cold start optimization config
        func._flypy_cold_start_config = {
            "optimize_cold_start": optimize_cold_start,
            "warmup_data": warmup_data or []
        }

        # Apply performance monitoring if enabled
        if enable_performance_monitoring:
            func = performance_monitor.monitor_function(enable_profiling=False)(func)

        # Create function definition
        func_def = FunctionDefinition(
            metadata=metadata,
            source_code=source_code,
        )

        # Analyze imports and dependencies
        _analyze_function_dependencies(func_def)

        # Store in registry
        _FUNCTION_REGISTRY[func_name] = func_def

        # Attach metadata to function for introspection
        func._flypy_metadata = metadata
        func._flypy_definition = func_def

        # Return the original function (no wrapper needed for FlyPy)
        return func

    return decorator


def input_schema(schema: Union[Dict[str, Any], Schema]) -> Callable:
    """
    Decorator to specify input schema for a FlyPy function.

    Args:
        schema: Input schema as dict or Schema object

    Returns:
        Decorator function

    Example:
        @flypy.input_schema({
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "age": {"type": "integer"}
            },
            "required": ["name"]
        })
        @flypy.function(name="process-user")
        def handler(event):
            return {"message": f"Hello {event['name']}!"}
    """

    def decorator(func: Callable) -> Callable:
        if not hasattr(func, '_flypy_definition'):
            raise ValueError(f"Function {func.__name__} must be decorated with @flypy.function first")

        func_def = func._flypy_definition

        if isinstance(schema, Schema):
            func_def.metadata.input_schema = schema.to_dict()
        else:
            func_def.metadata.input_schema = schema

        return func

    return decorator


def output_schema(schema: Union[Dict[str, Any], Schema]) -> Callable:
    """
    Decorator to specify output schema for a FlyPy function.

    Args:
        schema: Output schema as dict or Schema object

    Returns:
        Decorator function

    Example:
        @flypy.output_schema({
            "type": "object",
            "properties": {
                "result": {"type": "number"},
                "status": {"type": "string"}
            },
            "required": ["result"]
        })
        @flypy.function(name="calculate")
        def handler(event):
            return {"result": event["a"] + event["b"], "status": "success"}
    """

    def decorator(func: Callable) -> Callable:
        if not hasattr(func, '_flypy_definition'):
            raise ValueError(f"Function {func.__name__} must be decorated with @flypy.function first")

        func_def = func._flypy_definition

        if isinstance(schema, Schema):
            func_def.metadata.output_schema = schema.to_dict()
        else:
            func_def.metadata.output_schema = schema

        return func

    return decorator


def get_registered_functions() -> Dict[str, FunctionDefinition]:
    """Get all registered FlyPy functions."""
    return _FUNCTION_REGISTRY.copy()


def get_function_definition(name: str) -> Optional[FunctionDefinition]:
    """Get a specific function definition by name."""
    return _FUNCTION_REGISTRY.get(name)


def clear_registry() -> None:
    """Clear the function registry (useful for testing)."""
    _FUNCTION_REGISTRY.clear()


def _get_function_source(func: Callable) -> str:
    """Get the source code of a function."""
    try:
        return inspect.getsource(func)
    except (OSError, TypeError):
        # Fallback for functions where source is not available
        return f"def {func.__name__}{inspect.signature(func)}:\n    # Source not available\n    pass"


class ImportAnalyzer(ast.NodeVisitor):
    """AST visitor to analyze Python imports and dependencies."""

    def __init__(self):
        self.imports = []
        self.dependencies = set()

    def visit_Import(self, node: ast.Import) -> None:
        """Handle 'import module' statements."""
        # Debug: print node attributes
        # print(f"Import node attributes: {dir(node)}")

        for alias in node.names:  # Use 'names' instead of 'aliases'
            # Store the full import statement
            import_stmt = f"import {alias.name}"
            if alias.asname:
                import_stmt += f" as {alias.asname}"
            self.imports.append(import_stmt)

            # Extract the root module name for dependencies
            root_module = alias.name.split('.')[0]
            self.dependencies.add(root_module)

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        """Handle 'from module import ...' statements."""
        if node.module is None:
            return  # Relative import without module name

        # Build the import statement
        import_stmt = f"from {node.module} import "
        if node.names:
            names = []
            for alias in node.names:  # Use 'names' instead of 'aliases'
                if alias.name == '*':
                    names.append('*')
                elif alias.asname:
                    names.append(f"{alias.name} as {alias.asname}")
                else:
                    names.append(alias.name)
            import_stmt += ', '.join(names)
        else:
            import_stmt += '*'

        self.imports.append(import_stmt)

        # Extract the root module name for dependencies
        root_module = node.module.split('.')[0]
        self.dependencies.add(root_module)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        """Visit function definitions (but don't recurse into nested functions)."""
        # Only analyze the function body, not nested functions
        for item in node.body:
            self.visit(item)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        """Visit async function definitions."""
        for item in node.body:
            self.visit(item)


def _analyze_function_dependencies(func_def: FunctionDefinition) -> None:
    """Analyze function dependencies using AST parsing for accuracy."""
    try:
        import textwrap

        # Dedent the source code to handle indentation
        dedented_source = textwrap.dedent(func_def.source_code)

        # Try to parse the dedented source code
        tree = ast.parse(dedented_source)
        analyzer = ImportAnalyzer()
        analyzer.visit(tree)

        # Store results
        func_def.imports = analyzer.imports
        func_def.dependencies = list(analyzer.dependencies)

    except (SyntaxError, ValueError, IndentationError) as e:
        # Fallback to basic string analysis if AST parsing fails
        print(f"Warning: AST parsing failed for function dependencies ({e}), falling back to basic analysis", file=sys.stderr)
        _analyze_function_dependencies_basic(func_def)
    except Exception as e:
        # Fallback for any other parsing issues
        print(f"Warning: Dependency analysis failed ({e}), falling back to basic analysis", file=sys.stderr)
        _analyze_function_dependencies_basic(func_def)


def _analyze_function_dependencies_basic(func_def: FunctionDefinition) -> None:
    """Basic string-based dependency analysis (fallback)."""
    lines = func_def.source_code.split('\n')
    imports = []
    dependencies = []

    for line in lines:
        line = line.strip()
        if line.startswith('import '):
            imports.append(line)
            # Extract module name
            parts = line.split()
            if len(parts) >= 2:
                module = parts[1].split('.')[0]
                dependencies.append(module)
        elif line.startswith('from ') and ' import ' in line:
            imports.append(line)
            # Extract module name
            from_part = line.split(' import ')[0]
            module = from_part.replace('from ', '').split('.')[0]
            dependencies.append(module)

    func_def.imports = imports
    func_def.dependencies = list(set(dependencies))  # Remove duplicates


def auto_infer_schemas(func: Callable) -> Callable:
    """
    Automatically infer input and output schemas from type hints.

    This decorator analyzes the function's type hints and generates
    appropriate JSON schemas.

    Args:
        func: Function to analyze

    Returns:
        Decorated function with inferred schemas

    Example:
        @flypy.auto_infer_schemas
        @flypy.function(name="add-numbers")
        def add(a: int, b: int) -> Dict[str, int]:
            return {"result": a + b}
    """
    if not hasattr(func, '_flypy_definition'):
        raise ValueError(f"Function {func.__name__} must be decorated with @flypy.function first")

    func_def = func._flypy_definition

    # Infer schemas from type hints
    input_schema, output_schema = Schema.infer_from_function(func)

    # Apply schemas
    func_def.metadata.input_schema = input_schema.to_dict()
    func_def.metadata.output_schema = output_schema.to_dict()

    return func