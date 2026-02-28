"""
CLI commands for FlyPy.

Provides commands to build, deploy, and manage FlyPy functions.
"""

import argparse
import sys
import os
import json
import subprocess
import base64
import urllib.request
import urllib.error
import importlib.util
import threading
import time
import hashlib
import shutil
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from typing import List, Optional, Callable, Any, Dict

from .decorators import get_registered_functions, get_function_definition
from .types import BuildResult, ExecutionMode
from . import __version__


# ─────────────────────────────────────────────────────────────────────────────
# Config file support
# ─────────────────────────────────────────────────────────────────────────────

def _load_config(config_path: Optional[str] = None) -> Dict[str, Any]:
    """Load flypy.toml or flypy.json config from the project directory."""
    candidates = []
    if config_path:
        candidates.append(Path(config_path))
    else:
        cwd = Path.cwd()
        candidates = [
            cwd / "flypy.toml",
            cwd / "flypy.json",
            cwd / ".flypy.toml",
            cwd / ".flypy.json",
        ]

    for path in candidates:
        if path.exists():
            try:
                if path.suffix == ".toml":
                    try:
                        import tomllib  # Python 3.11+
                    except ImportError:
                        try:
                            import tomli as tomllib  # type: ignore
                        except ImportError:
                            return {}
                    with open(path, "rb") as f:
                        return tomllib.load(f)
                else:
                    with open(path) as f:
                        return json.load(f)
            except Exception:
                pass
    return {}


def _merge_config(args, config: Dict[str, Any], section: str) -> None:
    """Merge config file values into args (args take precedence)."""
    section_config = config.get(section, {})
    for key, value in section_config.items():
        attr = key.replace("-", "_")
        if not hasattr(args, attr) or getattr(args, attr) is None:
            setattr(args, attr, value)


# ─────────────────────────────────────────────────────────────────────────────
# Output helpers
# ─────────────────────────────────────────────────────────────────────────────

def _print(msg: str, quiet: bool = False, file=None) -> None:
    """Print unless quiet mode is active."""
    if not quiet:
        print(msg, file=file)


def _print_json(data: Any) -> None:
    """Print data as JSON."""
    print(json.dumps(data, indent=2))


def _hint(msg: str) -> str:
    """Format a hint message."""
    return f"   → {msg}"


# ─────────────────────────────────────────────────────────────────────────────
# Main entry point
# ─────────────────────────────────────────────────────────────────────────────

def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        description="FlyPy - Deterministic Python Compilation",
        prog="flypy"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"FlyPy {__version__}"
    )
    parser.add_argument(
        "--config",
        help="Path to flypy.toml or flypy.json config file"
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # ── init ──────────────────────────────────────────────────────────────────
    init_parser = subparsers.add_parser(
        "init",
        help="Scaffold a new FlyPy function file"
    )
    init_parser.add_argument(
        "name",
        help="Function name (e.g. my-function)"
    )
    init_parser.add_argument(
        "--output", "-o",
        default=".",
        help="Output directory (default: current directory)"
    )
    init_parser.add_argument(
        "--template", "-t",
        choices=["basic", "calculator", "data-transform", "api-call"],
        default="basic",
        help="Template to use (default: basic)"
    )
    init_parser.add_argument(
        "--force", "-f",
        action="store_true",
        help="Overwrite existing files"
    )

    # ── build ─────────────────────────────────────────────────────────────────
    build_parser = subparsers.add_parser(
        "build",
        help="Build FlyPy functions to WebAssembly"
    )
    build_parser.add_argument(
        "files",
        nargs="+",
        help="Python files containing FlyPy functions"
    )
    build_parser.add_argument(
        "--output", "-o",
        default=None,
        help="Output directory for artifacts (default: ./dist)"
    )
    build_parser.add_argument(
        "--mode",
        choices=["deterministic", "compatible"],
        default=None,
        help="Execution mode (default: deterministic)"
    )
    build_parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Verbose output"
    )
    build_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )
    build_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )
    build_parser.add_argument(
        "--go-binary",
        default=None,
        help="Path to FlyPy Go binary (auto-detected if not specified)"
    )
    build_parser.add_argument(
        "--optimize",
        action="store_true",
        default=None,
        help="Enable bundle size optimization (default: on)"
    )
    build_parser.add_argument(
        "--no-optimize",
        action="store_true",
        help="Disable bundle size optimization"
    )
    build_parser.add_argument(
        "--optimization-level",
        choices=["minimal", "balanced", "aggressive"],
        default=None,
        help="Optimization level (default: balanced)"
    )
    build_parser.add_argument(
        "--no-cold-start-optimize",
        action="store_true",
        help="Disable cold start optimization"
    )
    build_parser.add_argument(
        "--no-parallel",
        action="store_true",
        help="Disable parallel building"
    )
    build_parser.add_argument(
        "--no-incremental",
        action="store_true",
        help="Disable incremental builds"
    )
    build_parser.add_argument(
        "--max-workers",
        type=int,
        default=None,
        help="Maximum number of parallel workers (default: CPU count)"
    )

    # ── deploy ────────────────────────────────────────────────────────────────
    deploy_parser = subparsers.add_parser(
        "deploy",
        help="Deploy FlyPy functions to FunctionFly"
    )
    deploy_parser.add_argument(
        "artifact_dir",
        help="Directory containing built artifacts"
    )
    deploy_parser.add_argument(
        "--registry",
        default=None,
        help="FunctionFly registry URL (default: https://api.functionfly.com)"
    )
    deploy_parser.add_argument(
        "--token",
        default=None,
        help="Authentication token (or set FUNCTIONFLY_TOKEN env var)"
    )
    deploy_parser.add_argument(
        "--app-id",
        required=True,
        help="FunctionFly app ID to deploy to"
    )
    deploy_parser.add_argument(
        "--provider",
        default=None,
        choices=["cloudflare", "vercel", "fly", "deno"],
        help="Cloud provider for deployment (default: cloudflare)"
    )
    deploy_parser.add_argument(
        "--region",
        default=None,
        help="Deployment region (default: us-east-1)"
    )
    deploy_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )
    deploy_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )

    # ── list ──────────────────────────────────────────────────────────────────
    list_parser = subparsers.add_parser(
        "list",
        help="List registered FlyPy functions"
    )
    list_parser.add_argument(
        "files",
        nargs="*",
        help="Python files to scan (optional, defaults to current directory)"
    )
    list_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )
    list_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )

    # ── local ─────────────────────────────────────────────────────────────────
    local_parser = subparsers.add_parser(
        "local",
        help="Run FlyPy functions locally for testing"
    )
    local_parser.add_argument(
        "file",
        help="Python file containing the function"
    )
    local_parser.add_argument(
        "function",
        help="Function name to run"
    )
    local_parser.add_argument(
        "--port", "-p",
        type=int,
        default=None,
        help="Port to run local server on (default: 8080)"
    )
    local_parser.add_argument(
        "--watch", "-w",
        action="store_true",
        help="Watch for file changes and auto-reload"
    )
    local_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )

    # ── verify ────────────────────────────────────────────────────────────────
    verify_parser = subparsers.add_parser(
        "verify",
        help="Verify determinism of built artifacts"
    )
    verify_parser.add_argument(
        "artifact_dir",
        help="Directory containing built artifacts"
    )
    verify_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )
    verify_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )

    # ── inspect ───────────────────────────────────────────────────────────────
    inspect_parser = subparsers.add_parser(
        "inspect",
        help="Show detailed information about a built artifact"
    )
    inspect_parser.add_argument(
        "artifact_dir",
        help="Directory containing built artifacts"
    )
    inspect_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )

    # ── clean ─────────────────────────────────────────────────────────────────
    clean_parser = subparsers.add_parser(
        "clean",
        help="Remove build artifacts and cache"
    )
    clean_parser.add_argument(
        "--output", "-o",
        default=None,
        help="Output directory to clean (default: ./dist)"
    )
    clean_parser.add_argument(
        "--cache",
        action="store_true",
        help="Also clear the build cache"
    )
    clean_parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be deleted without deleting"
    )
    clean_parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Suppress all output except errors"
    )

    # ── monitor ───────────────────────────────────────────────────────────────
    monitor_parser = subparsers.add_parser(
        "monitor",
        help="Performance monitoring and profiling"
    )
    monitor_parser.add_argument(
        "--start",
        action="store_true",
        help="Start performance monitoring"
    )
    monitor_parser.add_argument(
        "--stop",
        action="store_true",
        help="Stop performance monitoring"
    )
    monitor_parser.add_argument(
        "--report",
        action="store_true",
        help="Generate performance report"
    )
    monitor_parser.add_argument(
        "--dashboard",
        action="store_true",
        help="Start performance dashboard"
    )
    monitor_parser.add_argument(
        "--alerts",
        action="store_true",
        help="Check for performance alerts"
    )
    monitor_parser.add_argument(
        "--interval",
        type=int,
        default=60,
        help="Monitoring interval in seconds (default: 60)"
    )
    monitor_parser.add_argument(
        "--host",
        default="localhost",
        help="Dashboard host (default: localhost)"
    )
    monitor_parser.add_argument(
        "--port",
        type=int,
        default=8081,
        help="Dashboard port (default: 8081)"
    )
    monitor_parser.add_argument(
        "--json",
        action="store_true",
        help="Output results as JSON"
    )

    # ── completion ────────────────────────────────────────────────────────────
    completion_parser = subparsers.add_parser(
        "completion",
        help="Generate shell completion scripts"
    )
    completion_parser.add_argument(
        "shell",
        choices=["bash", "zsh", "fish"],
        help="Shell to generate completion for"
    )

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    # Load config file
    config = _load_config(getattr(args, "config", None))

    verbose = getattr(args, "verbose", False)

    try:
        if args.command == "init":
            return init_command(args, config)
        elif args.command == "build":
            return build_command(args, config)
        elif args.command == "deploy":
            return deploy_command(args, config)
        elif args.command == "list":
            return list_command(args, config)
        elif args.command == "local":
            return local_command(args, config)
        elif args.command == "verify":
            return verify_command(args, config)
        elif args.command == "inspect":
            return inspect_command(args)
        elif args.command == "clean":
            return clean_command(args, config)
        elif args.command == "monitor":
            return monitor_command(args)
        elif args.command == "completion":
            return completion_command(args)
        else:
            parser.print_help()
            return 1
    except KeyboardInterrupt:
        print("\n🛑 Interrupted", file=sys.stderr)
        return 130
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        if verbose:
            import traceback
            traceback.print_exc()
        return 1


# ─────────────────────────────────────────────────────────────────────────────
# init command
# ─────────────────────────────────────────────────────────────────────────────

_TEMPLATES: Dict[str, str] = {
    "basic": '''\
import flypy


@flypy.function(
    name="{name}",
    version="1.0.0",
    description="A FlyPy function",
    deterministic=True,
    idempotent=True,
)
def handler(event):
    """Process the input event and return a result."""
    return {{"result": event}}
''',
    "calculator": '''\
import flypy
from typing import Dict, Any


@flypy.function(
    name="{name}",
    version="1.0.0",
    description="Perform arithmetic calculations",
    deterministic=True,
    idempotent=True,
    pure=True,
)
def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Calculate the result of an arithmetic expression.

    Input:
        a (number): First operand
        b (number): Second operand
        op (str): Operation: "add", "sub", "mul", "div"

    Output:
        result (number): Calculation result
    """
    a = event.get("a", 0)
    b = event.get("b", 0)
    op = event.get("op", "add")

    if op == "add":
        result = a + b
    elif op == "sub":
        result = a - b
    elif op == "mul":
        result = a * b
    elif op == "div":
        if b == 0:
            raise ValueError("Division by zero")
        result = a / b
    else:
        raise ValueError(f"Unknown operation: {{op}}")

    return {{"result": result, "op": op, "a": a, "b": b}}
''',
    "data-transform": '''\
import flypy
from typing import Dict, Any, List


@flypy.function(
    name="{name}",
    version="1.0.0",
    description="Transform a list of data records",
    deterministic=True,
    idempotent=True,
)
def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Transform input records.

    Input:
        items (list): List of records to transform
        config (dict): Transformation configuration

    Output:
        results (list): Transformed records
        count (int): Number of records processed
    """
    items: List[Dict] = event.get("items", [])
    config: Dict = event.get("config", {{}})

    results = []
    for item in items:
        transformed = {{
            "id": item.get("id"),
            "processed": True,
            **item,
        }}
        results.append(transformed)

    return {{"results": results, "count": len(results)}}
''',
    "api-call": '''\
import flypy
from typing import Dict, Any


@flypy.function(
    name="{name}",
    version="1.0.0",
    description="Call an external API",
    deterministic=False,
    capabilities=["network"],
)
def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Fetch data from an external API.

    Input:
        url (str): API endpoint URL
        method (str): HTTP method (default: GET)
        body (dict): Request body (optional)

    Output:
        status (int): HTTP status code
        data (any): Response data
    """
    import urllib.request
    import json

    url = event.get("url")
    if not url:
        raise ValueError("url is required")

    method = event.get("method", "GET").upper()
    body = event.get("body")

    data = json.dumps(body).encode() if body else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")

    with urllib.request.urlopen(req) as resp:
        response_data = json.loads(resp.read().decode())
        return {{"status": resp.status, "data": response_data}}
''',
}


def init_command(args, config: Dict[str, Any]) -> int:
    """Scaffold a new FlyPy function file."""
    name = args.name
    output_dir = Path(args.output)
    template_name = args.template
    force = args.force

    # Sanitize name for use as Python identifier
    py_name = name.replace("-", "_")
    file_name = f"{py_name}.py"
    output_file = output_dir / file_name

    if output_file.exists() and not force:
        print(f"Error: {output_file} already exists", file=sys.stderr)
        print(_hint("Use --force to overwrite"), file=sys.stderr)
        return 1

    output_dir.mkdir(parents=True, exist_ok=True)

    template = _TEMPLATES.get(template_name, _TEMPLATES["basic"])
    content = template.format(name=name)

    with open(output_file, "w") as f:
        f.write(content)

    print(f"✅ Created {output_file}")
    print()
    print("Next steps:")
    print(f"  1. Edit {output_file} to implement your function")
    print(f"  2. Build:  flypy build {output_file}")
    print(f"  3. Test:   flypy local {output_file} {name}")
    print(f"  4. Verify: flypy verify ./dist/{name}")
    print(f"  5. Deploy: flypy deploy ./dist/{name} --app-id <your-app-id>")
    return 0


# ─────────────────────────────────────────────────────────────────────────────
# build command
# ─────────────────────────────────────────────────────────────────────────────

def build_command(args, config: Dict[str, Any]) -> int:
    """Build FlyPy functions."""
    # Merge config defaults
    _merge_config(args, config, "build")

    # Apply defaults after config merge
    output = getattr(args, "output", None) or "./dist"
    mode = getattr(args, "mode", None) or "deterministic"
    verbose = getattr(args, "verbose", False)
    quiet = getattr(args, "quiet", False)
    as_json = getattr(args, "json", False)
    go_binary = getattr(args, "go_binary", None)
    optimization_level = getattr(args, "optimization_level", None) or "balanced"
    max_workers = getattr(args, "max_workers", None)

    # Resolve boolean flags
    optimize_bundle = not getattr(args, "no_optimize", False)
    optimize_cold_start = not getattr(args, "no_cold_start_optimize", False)
    enable_parallel = not getattr(args, "no_parallel", False)
    enable_incremental = not getattr(args, "no_incremental", False)

    if not quiet and not as_json:
        print("🔨 Building FlyPy functions...")

    # Load Python files to register functions
    for file_path in args.files:
        if not os.path.exists(file_path):
            print(f"Error: File {file_path} not found", file=sys.stderr)
            print(_hint(f"Check the path and try again"), file=sys.stderr)
            return 1
        if not quiet and not as_json:
            print(f"Loading {file_path}...")
        _load_python_file(file_path)

    # Get registered functions
    functions = get_registered_functions()
    if not functions:
        print("No FlyPy functions found in the specified files", file=sys.stderr)
        print(_hint("Make sure your functions use the @flypy.function decorator"), file=sys.stderr)
        return 1

    if not quiet and not as_json:
        print(f"Found {len(functions)} function(s): {', '.join(functions.keys())}")

    # Create output directory
    output_dir = Path(output)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Find Go binary
    resolved_go_binary = go_binary or _find_go_binary()

    # Build results tracking
    json_results = []
    success_count = 0

    if len(functions) > 1 and (enable_parallel or enable_incremental):
        from flypy.build import build_all

        results = build_all(
            output_dir=str(output_dir),
            mode=mode,
            verbose=verbose,
            go_binary=resolved_go_binary,
            optimize_bundle=optimize_bundle,
            optimization_level=optimization_level,
            optimize_cold_start=optimize_cold_start,
            enable_parallel=enable_parallel,
            enable_incremental=enable_incremental,
            max_parallel_workers=max_workers
        )

        for result in results:
            if result.success:
                success_count += 1
                if not quiet and not as_json:
                    print(f"✅ Built {result.function_name} successfully")
                    if verbose:
                        print(f"   Output: {result.output_dir}")
                        if result.wasm_size_bytes:
                            print(f"   WASM size: {result.wasm_size_bytes:,} bytes")
                        if result.build_time_ms:
                            print(f"   Build time: {result.build_time_ms}ms")
            else:
                if not quiet and not as_json:
                    print(f"❌ Failed to build {result.function_name}")
                    for error in result.errors:
                        print(f"   Error: {error}")
                else:
                    for error in result.errors:
                        print(f"Error [{result.function_name}]: {error}", file=sys.stderr)

            json_results.append({
                "function_name": result.function_name,
                "success": result.success,
                "output_dir": result.output_dir,
                "wasm_size_bytes": result.wasm_size_bytes,
                "build_time_ms": result.build_time_ms,
                "errors": result.errors,
                "warnings": result.warnings,
            })
    else:
        for func_name, func_def in functions.items():
            if not quiet and not as_json:
                print(f"\nBuilding function: {func_name}")

            try:
                result = _build_function(
                    resolved_go_binary, func_def, output_dir, mode, verbose,
                    optimize_bundle, optimization_level, optimize_cold_start
                )
                if result.success:
                    success_count += 1
                    if not quiet and not as_json:
                        print(f"✅ Built {func_name} successfully")
                        if verbose:
                            print(f"   Output: {result.output_dir}")
                            if result.wasm_size_bytes:
                                print(f"   WASM size: {result.wasm_size_bytes:,} bytes")
                            if result.build_time_ms:
                                print(f"   Build time: {result.build_time_ms}ms")
                else:
                    if not quiet and not as_json:
                        print(f"❌ Failed to build {func_name}")
                        for error in result.errors:
                            print(f"   Error: {error}")
                    else:
                        for error in result.errors:
                            print(f"Error [{func_name}]: {error}", file=sys.stderr)

                json_results.append({
                    "function_name": result.function_name,
                    "success": result.success,
                    "output_dir": result.output_dir,
                    "wasm_size_bytes": result.wasm_size_bytes,
                    "build_time_ms": result.build_time_ms,
                    "errors": result.errors,
                    "warnings": result.warnings,
                })

            except Exception as e:
                if not quiet and not as_json:
                    print(f"❌ Build failed for {func_name}: {e}")
                else:
                    print(f"Error [{func_name}]: {e}", file=sys.stderr)
                if verbose:
                    import traceback
                    traceback.print_exc()
                json_results.append({
                    "function_name": func_name,
                    "success": False,
                    "output_dir": str(output_dir / func_name),
                    "errors": [str(e)],
                })

    total = len(functions)
    overall_success = success_count == total

    if as_json:
        _print_json({
            "success": overall_success,
            "built": success_count,
            "total": total,
            "functions": json_results,
        })
    elif not quiet:
        if overall_success:
            print(f"\n🎉 Successfully built all {success_count} function(s)!")
        else:
            print(f"\n⚠️  Built {success_count}/{total} function(s)")

    return 0 if overall_success else 1


# ─────────────────────────────────────────────────────────────────────────────
# deploy command
# ─────────────────────────────────────────────────────────────────────────────

def deploy_command(args, config: Dict[str, Any]) -> int:
    """Deploy FlyPy functions."""
    _merge_config(args, config, "deploy")

    registry = getattr(args, "registry", None) or "https://api.functionfly.com"
    provider = getattr(args, "provider", None) or "cloudflare"
    region = getattr(args, "region", None) or "us-east-1"
    quiet = getattr(args, "quiet", False)
    as_json = getattr(args, "json", False)

    if not quiet and not as_json:
        print("📦 Deploying FlyPy functions...")

    artifact_dir = Path(args.artifact_dir)
    if not artifact_dir.exists():
        print(f"Error: Artifact directory {artifact_dir} not found", file=sys.stderr)
        print(_hint(f"Run 'flypy build' first to create artifacts"), file=sys.stderr)
        return 1

    # Check for required files
    manifest_file = artifact_dir / "manifest.json"
    wasm_file = artifact_dir / "state_transition.wasm"

    if not manifest_file.exists():
        print(f"Error: manifest.json not found in {artifact_dir}", file=sys.stderr)
        print(_hint("The artifact directory may be incomplete — try rebuilding"), file=sys.stderr)
        return 1

    if not wasm_file.exists():
        print(f"Error: state_transition.wasm not found in {artifact_dir}", file=sys.stderr)
        print(_hint("The artifact directory may be incomplete — try rebuilding"), file=sys.stderr)
        return 1

    # Load manifest
    try:
        with open(manifest_file) as f:
            manifest = json.load(f)
        func_name = manifest.get("name", "unknown")
    except Exception as e:
        print(f"Error reading manifest: {e}", file=sys.stderr)
        return 1

    if not quiet and not as_json:
        print(f"Deploying function: {func_name}")

    # Resolve token: arg → env var
    token = args.token or os.environ.get("FUNCTIONFLY_TOKEN")
    if not token:
        print("Error: Authentication token is required", file=sys.stderr)
        print(_hint("Use --token <token> or set the FUNCTIONFLY_TOKEN environment variable"), file=sys.stderr)
        print(_hint("Get your token at https://app.functionfly.com/settings/tokens"), file=sys.stderr)
        return 1

    # Read and encode WASM artifact
    try:
        with open(wasm_file, "rb") as f:
            wasm_data = f.read()
        artifact_b64 = base64.b64encode(wasm_data).decode("utf-8")
    except Exception as e:
        print(f"Error reading WASM file: {e}", file=sys.stderr)
        return 1

    # Prepare deployment request
    deploy_data = {
        "provider": provider,
        "region": region,
        "artifact": artifact_b64,
        "routes": manifest.get("routes", []),
        "env_vars": manifest.get("env_vars", {}),
        "secrets": manifest.get("secrets", {}),
        "provider_config": manifest.get("provider_config", {})
    }

    # Make deployment request
    try:
        url = f"{registry}/apps/{args.app_id}/deploy"
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}"
        }

        data = json.dumps(deploy_data).encode("utf-8")
        req = urllib.request.Request(url, data=data, headers=headers, method="POST")

        if not quiet and not as_json:
            print(f"🚀 Deploying to {provider} in {region}...")

        with urllib.request.urlopen(req) as response:
            if response.status == 201:
                result = json.loads(response.read().decode("utf-8"))
                deployment_id = result.get("id", "unknown")

                if as_json:
                    _print_json({
                        "success": True,
                        "deployment_id": deployment_id,
                        "status": result.get("status", "unknown"),
                        "url": result.get("url"),
                        "function": func_name,
                    })
                elif not quiet:
                    print(f"✅ Deployment successful!")
                    print(f"   Deployment ID: {deployment_id}")
                    print(f"   Status: {result.get('status', 'unknown')}")
                    if "url" in result:
                        print(f"   URL: {result['url']}")
                return 0
            else:
                error_data = json.loads(response.read().decode("utf-8"))
                msg = error_data.get("error", "Unknown error")
                print(f"❌ Deployment failed: {msg}", file=sys.stderr)
                return 1

    except urllib.error.HTTPError as e:
        try:
            error_data = json.loads(e.read().decode("utf-8"))
            msg = error_data.get("error", str(e))
        except Exception:
            msg = str(e)
        print(f"❌ Deployment failed (HTTP {e.code}): {msg}", file=sys.stderr)
        if e.code == 401:
            print(_hint("Your token may be expired — get a new one at https://app.functionfly.com/settings/tokens"), file=sys.stderr)
        elif e.code == 409:
            print(_hint("This version may already be deployed — try bumping the version"), file=sys.stderr)
        return 1
    except urllib.error.URLError as e:
        print(f"❌ Network error: {e}", file=sys.stderr)
        print(_hint("Check your internet connection and try again"), file=sys.stderr)
        return 1
    except json.JSONDecodeError as e:
        print(f"❌ Invalid response from server: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"❌ Unexpected error during deployment: {e}", file=sys.stderr)
        return 1


# ─────────────────────────────────────────────────────────────────────────────
# list command
# ─────────────────────────────────────────────────────────────────────────────

def list_command(args, config: Dict[str, Any]) -> int:
    """List registered FlyPy functions."""
    quiet = getattr(args, "quiet", False)
    as_json = getattr(args, "json", False)
    files_to_scan = args.files or ["."]

    for file_path in files_to_scan:
        if os.path.isdir(file_path):
            python_files = list(Path(file_path).rglob("*.py"))
            for py_file in python_files:
                _load_python_file(str(py_file))
        else:
            _load_python_file(file_path)

    functions = get_registered_functions()

    if not functions:
        if as_json:
            _print_json({"functions": []})
        elif not quiet:
            print("No FlyPy functions found")
        return 0

    if as_json:
        fn_list = []
        for name, func_def in functions.items():
            m = func_def.metadata
            fn_list.append({
                "name": name,
                "version": m.version,
                "description": m.description,
                "deterministic": m.deterministic,
                "idempotent": m.idempotent,
                "pure": m.pure,
                "cache_ttl": m.cache_ttl,
                "capabilities": m.capabilities,
                "source_file": m.source_file,
            })
        _print_json({"functions": fn_list})
        return 0

    if not quiet:
        print(f"Found {len(functions)} FlyPy function(s):")
        print()

        for name, func_def in functions.items():
            metadata = func_def.metadata
            print(f"📦 {name} (v{metadata.version})")
            if metadata.description:
                print(f"   Description: {metadata.description}")
            print(f"   Deterministic: {'✅' if metadata.deterministic else '❌'}")
            print(f"   Idempotent:    {'✅' if metadata.idempotent else '❌'}")
            print(f"   Pure:          {'✅' if metadata.pure else '❌'}")
            if metadata.capabilities:
                print(f"   Capabilities:  {', '.join(metadata.capabilities)}")
            if metadata.cache_ttl:
                print(f"   Cache TTL:     {metadata.cache_ttl}s")
            if metadata.source_file:
                print(f"   Source:        {metadata.source_file}")
            print()

    return 0


# ─────────────────────────────────────────────────────────────────────────────
# local command (with hot reload)
# ─────────────────────────────────────────────────────────────────────────────

class _ReloadableServer:
    """HTTP server wrapper that supports hot reload."""

    def __init__(self, file_path: str, function_name: str, port: int, quiet: bool):
        self.file_path = file_path
        self.function_name = function_name
        self.port = port
        self.quiet = quiet
        self._server: Optional[HTTPServer] = None
        self._server_thread: Optional[threading.Thread] = None
        self._function: Optional[Callable] = None
        self._lock = threading.Lock()

    def _load(self) -> bool:
        module = _load_python_file(self.file_path)
        if not module:
            return False
        func_def = get_function_definition(self.function_name)
        if not func_def:
            return False
        function = _get_function_from_module(module, self.function_name)
        if not function:
            return False
        self._function = function
        return True

    def start(self) -> bool:
        if not self._load():
            return False
        self._start_server()
        return True

    def reload(self) -> None:
        if not self.quiet:
            print("♻️  Reloading...")
        with self._lock:
            if self._load():
                if not self.quiet:
                    print(f"✅ Reloaded {self.function_name}")
            else:
                if not self.quiet:
                    print(f"⚠️  Reload failed — keeping previous version", file=sys.stderr)

    def _start_server(self) -> None:
        server_ref = self

        def handler_factory(*args, **kwargs):
            return FunctionRequestHandler(server_ref, *args, **kwargs)

        self._server = HTTPServer(("localhost", self.port), handler_factory)
        self._server_thread = threading.Thread(
            target=self._server.serve_forever,
            daemon=True
        )
        self._server_thread.start()

    def stop(self) -> None:
        if self._server:
            self._server.shutdown()


class FunctionRequestHandler(BaseHTTPRequestHandler):
    """HTTP request handler for FlyPy function execution."""

    def __init__(self, server_ref: "_ReloadableServer", *args, **kwargs):
        self.server_ref = server_ref
        super().__init__(*args, **kwargs)

    def do_POST(self) -> None:
        """Handle POST requests by executing the function."""
        try:
            content_length = int(self.headers.get("Content-Length", 0))
            post_data = self.rfile.read(content_length) if content_length else b"{}"

            try:
                input_data = json.loads(post_data.decode("utf-8"))
            except json.JSONDecodeError:
                self.send_error(400, "Invalid JSON in request body")
                return

            with self.server_ref._lock:
                function = self.server_ref._function

            if not function:
                self.send_error(503, "Function not loaded")
                return

            try:
                result = function(input_data)
            except Exception as e:
                self.send_error(500, f"Function execution error: {str(e)}")
                return

            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(result).encode("utf-8"))

        except Exception as e:
            self.send_error(500, f"Server error: {str(e)}")

    def do_GET(self) -> None:
        """Handle GET requests with function info."""
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()

        with self.server_ref._lock:
            function = self.server_ref._function

        metadata = getattr(function, "_flypy_metadata", None) if function else None
        info = {
            "status": "running",
            "function": self.server_ref.function_name,
            "description": metadata.description if metadata else "",
            "version": metadata.version if metadata else "unknown",
        }
        self.wfile.write(json.dumps(info).encode("utf-8"))

    def log_message(self, format, *args):
        if not self.server_ref.quiet:
            ts = time.strftime("%H:%M:%S")
            print(f"[{ts}] {format % args}")


def local_command(args, config: Dict[str, Any]) -> int:
    """Run FlyPy functions locally."""
    _merge_config(args, config, "local")

    port = getattr(args, "port", None) or 8080
    watch = getattr(args, "watch", False)
    quiet = getattr(args, "quiet", False)

    if not os.path.exists(args.file):
        print(f"Error: File {args.file} not found", file=sys.stderr)
        return 1

    server = _ReloadableServer(args.file, args.function, port, quiet)

    if not server.start():
        print(f"Error: Could not load function '{args.function}' from {args.file}", file=sys.stderr)
        print(_hint(f"Make sure the function is decorated with @flypy.function(name='{args.function}')"), file=sys.stderr)
        return 1

    if not quiet:
        print(f"🏠 FlyPy local server running")
        print(f"   Function: {args.function}")
        print(f"   URL:      http://localhost:{port}")
        print(f"   GET  /    → function info")
        print(f"   POST /    → execute function (JSON body)")
        if watch:
            print(f"   Watching: {args.file}")
        print("   Press Ctrl+C to stop")

    if watch:
        _watch_and_reload(args.file, server, quiet)
    else:
        try:
            while True:
                time.sleep(0.5)
        except KeyboardInterrupt:
            pass

    if not quiet:
        print("\n🛑 Server stopped")
    server.stop()
    return 0


def _watch_and_reload(file_path: str, server: _ReloadableServer, quiet: bool) -> None:
    """Watch a file for changes and trigger reload."""
    try:
        from watchdog.observers import Observer  # type: ignore
        from watchdog.events import FileSystemEventHandler  # type: ignore

        class _Handler(FileSystemEventHandler):
            def on_modified(self, event):
                if not event.is_directory and event.src_path.endswith(file_path.lstrip("./")):
                    server.reload()

        observer = Observer()
        watch_dir = str(Path(file_path).parent.resolve())
        observer.schedule(_Handler(), watch_dir, recursive=False)
        observer.start()

        try:
            while True:
                time.sleep(0.5)
        except KeyboardInterrupt:
            pass
        finally:
            observer.stop()
            observer.join()

    except ImportError:
        # Fallback: poll for mtime changes
        if not quiet:
            print("   (Install 'watchdog' for better file watching: pip install watchdog)")

        last_mtime = os.path.getmtime(file_path)
        try:
            while True:
                time.sleep(1)
                try:
                    mtime = os.path.getmtime(file_path)
                    if mtime != last_mtime:
                        last_mtime = mtime
                        server.reload()
                except OSError:
                    pass
        except KeyboardInterrupt:
            pass


# ─────────────────────────────────────────────────────────────────────────────
# verify command
# ─────────────────────────────────────────────────────────────────────────────

def verify_command(args, config: Dict[str, Any]) -> int:
    """Verify determinism of artifacts."""
    quiet = getattr(args, "quiet", False)
    as_json = getattr(args, "json", False)

    if not quiet and not as_json:
        print("🔍 Verifying artifact determinism...")

    artifact_dir = Path(args.artifact_dir)
    if not artifact_dir.exists():
        print(f"Error: Artifact directory {artifact_dir} not found", file=sys.stderr)
        print(_hint("Run 'flypy build' first to create artifacts"), file=sys.stderr)
        return 1

    manifest_file = artifact_dir / "manifest.json"
    hash_file = artifact_dir / "determinism.hash"
    sig_file = artifact_dir / "signature.sig"

    for required, name in [(manifest_file, "manifest.json"), (hash_file, "determinism.hash"), (sig_file, "signature.sig")]:
        if not required.exists():
            print(f"Error: {name} not found in {artifact_dir}", file=sys.stderr)
            print(_hint("The artifact may be incomplete — try rebuilding"), file=sys.stderr)
            return 1

    checks = {}

    try:
        with open(manifest_file) as f:
            manifest = json.load(f)
        with open(hash_file) as f:
            stored_hash = f.read().strip()

        func_name = manifest.get("name", "unknown")
        if not quiet and not as_json:
            print(f"Verifying function: {func_name}")

        # Verify determinism hash
        wasm_file = artifact_dir / "state_transition.wasm"
        if not wasm_file.exists():
            print(f"Error: state_transition.wasm not found in {artifact_dir}", file=sys.stderr)
            return 1

        with open(wasm_file, "rb") as f:
            wasm_data = f.read()

        computed_hash = hashlib.sha256(wasm_data).hexdigest()

        if computed_hash != stored_hash:
            checks["hash"] = {"passed": False, "message": "Hash mismatch — artifact may have been tampered with"}
            if not quiet and not as_json:
                print(f"❌ Hash verification failed!")
                print(f"   Expected: {stored_hash}")
                print(f"   Computed: {computed_hash}")
        else:
            checks["hash"] = {"passed": True, "message": "Determinism hash verified"}
            if not quiet and not as_json:
                print("✅ Determinism hash verified")

        # Verify signature
        with open(sig_file, "rb") as f:
            signature_data = f.read()

        if len(signature_data) == 0:
            checks["signature"] = {"passed": True, "message": "No signature data (skipped)"}
            if not quiet and not as_json:
                print("⚠️  No signature data found (verification skipped)")
        elif len(signature_data) != 64:
            checks["signature"] = {"passed": False, "message": f"Invalid signature format ({len(signature_data)} bytes, expected 64)"}
            if not quiet and not as_json:
                print(f"❌ Invalid signature format", file=sys.stderr)
        else:
            checks["signature"] = {"passed": True, "message": "Signature format verified (Ed25519)"}
            if not quiet and not as_json:
                print("✅ Signature format verified")

        # Validate WASM magic bytes
        if len(wasm_data) < 8:
            checks["wasm"] = {"passed": False, "message": "WASM file too small"}
            if not quiet and not as_json:
                print("❌ WASM file too small", file=sys.stderr)
        elif wasm_data[:4] != b"\x00asm":
            checks["wasm"] = {"passed": False, "message": "Invalid WASM magic bytes"}
            if not quiet and not as_json:
                print("❌ Invalid WASM magic bytes", file=sys.stderr)
        elif wasm_data[4:8] != b"\x01\x00\x00\x00":
            checks["wasm"] = {"passed": False, "message": "Invalid WASM version"}
            if not quiet and not as_json:
                print("❌ Invalid WASM version", file=sys.stderr)
        else:
            checks["wasm"] = {"passed": True, "message": "WASM module structure valid"}
            if not quiet and not as_json:
                print("✅ WASM module structure verified")

        all_passed = all(c["passed"] for c in checks.values())

        if as_json:
            _print_json({
                "success": all_passed,
                "function": func_name,
                "artifact_dir": str(artifact_dir),
                "checks": checks,
                "wasm_size_bytes": len(wasm_data),
            })
        elif not quiet:
            if all_passed:
                print(f"\n🎉 All verifications passed!")
                print(f"   Function '{func_name}' artifacts are valid and deterministic")
            else:
                failed = [k for k, v in checks.items() if not v["passed"]]
                print(f"\n❌ Verification failed: {', '.join(failed)}")

        return 0 if all_passed else 1

    except Exception as e:
        print(f"Error during verification: {e}", file=sys.stderr)
        return 1


# ─────────────────────────────────────────────────────────────────────────────
# inspect command
# ─────────────────────────────────────────────────────────────────────────────

def inspect_command(args) -> int:
    """Show detailed information about a built artifact."""
    as_json = getattr(args, "json", False)
    artifact_dir = Path(args.artifact_dir)

    if not artifact_dir.exists():
        print(f"Error: Artifact directory {artifact_dir} not found", file=sys.stderr)
        return 1

    info: Dict[str, Any] = {"artifact_dir": str(artifact_dir)}

    # Manifest
    manifest_file = artifact_dir / "manifest.json"
    if manifest_file.exists():
        with open(manifest_file) as f:
            info["manifest"] = json.load(f)

    # Function metadata
    metadata_file = artifact_dir / "function_metadata.json"
    if metadata_file.exists():
        with open(metadata_file) as f:
            info["function_metadata"] = json.load(f)

    # WASM stats
    wasm_file = artifact_dir / "state_transition.wasm"
    if wasm_file.exists():
        stat = wasm_file.stat()
        with open(wasm_file, "rb") as f:
            wasm_data = f.read()
        info["wasm"] = {
            "file": str(wasm_file),
            "size_bytes": stat.st_size,
            "size_kb": round(stat.st_size / 1024, 2),
            "sha256": hashlib.sha256(wasm_data).hexdigest(),
            "valid_magic": wasm_data[:4] == b"\x00asm",
        }

    # Hash
    hash_file = artifact_dir / "determinism.hash"
    if hash_file.exists():
        with open(hash_file) as f:
            info["determinism_hash"] = f.read().strip()

    # Signature
    sig_file = artifact_dir / "signature.sig"
    if sig_file.exists():
        sig_data = sig_file.read_bytes()
        info["signature"] = {
            "present": len(sig_data) > 0,
            "size_bytes": len(sig_data),
            "valid_format": len(sig_data) == 64,
        }

    # Source
    source_file = artifact_dir / "function.py"
    if source_file.exists():
        stat = source_file.stat()
        info["source"] = {
            "file": str(source_file),
            "size_bytes": stat.st_size,
        }

    if as_json:
        _print_json(info)
        return 0

    # Human-readable output
    func_name = info.get("manifest", {}).get("name", artifact_dir.name)
    print(f"📦 Artifact: {func_name}")
    print(f"   Directory: {artifact_dir}")
    print()

    if "manifest" in info:
        m = info["manifest"]
        print("Manifest:")
        print(f"  Name:        {m.get('name', 'unknown')}")
        print(f"  Version:     {m.get('version', 'unknown')}")
        print(f"  Runtime:     {m.get('runtime', 'unknown')}")
        print()

    if "wasm" in info:
        w = info["wasm"]
        print("WebAssembly:")
        print(f"  Size:        {w['size_bytes']:,} bytes ({w['size_kb']} KB)")
        print(f"  SHA-256:     {w['sha256']}")
        print(f"  Valid:       {'✅' if w['valid_magic'] else '❌'}")
        print()

    if "determinism_hash" in info:
        print(f"Determinism Hash: {info['determinism_hash']}")

    if "signature" in info:
        s = info["signature"]
        status = "✅ present" if s["present"] and s["valid_format"] else "⚠️  missing or invalid"
        print(f"Signature:        {status}")

    return 0


# ─────────────────────────────────────────────────────────────────────────────
# clean command
# ─────────────────────────────────────────────────────────────────────────────

def clean_command(args, config: Dict[str, Any]) -> int:
    """Remove build artifacts and cache."""
    _merge_config(args, config, "build")

    output = getattr(args, "output", None) or "./dist"
    also_cache = getattr(args, "cache", False)
    dry_run = getattr(args, "dry_run", False)
    quiet = getattr(args, "quiet", False)

    output_dir = Path(output)
    removed = []

    if output_dir.exists():
        if dry_run:
            if not quiet:
                print(f"Would remove: {output_dir}")
            for item in output_dir.rglob("*"):
                if not quiet:
                    print(f"  {item}")
        else:
            shutil.rmtree(output_dir)
            removed.append(str(output_dir))
            if not quiet:
                print(f"🗑️  Removed {output_dir}")
    else:
        if not quiet:
            print(f"Nothing to clean: {output_dir} does not exist")

    if also_cache:
        try:
            from flypy.build_optimizer import clear_build_cache
            if dry_run:
                if not quiet:
                    print("Would clear build cache")
            else:
                clear_build_cache()
                removed.append("build cache")
                if not quiet:
                    print("🗑️  Cleared build cache")
        except Exception as e:
            if not quiet:
                print(f"Warning: Could not clear build cache: {e}", file=sys.stderr)

    if not quiet and not dry_run:
        if removed:
            print(f"✅ Clean complete")
        else:
            print("Nothing was removed")

    return 0


# ─────────────────────────────────────────────────────────────────────────────
# monitor command
# ─────────────────────────────────────────────────────────────────────────────

def monitor_command(args) -> int:
    """Performance monitoring commands."""
    as_json = getattr(args, "json", False)

    from flypy.performance_monitor import (
        start_performance_monitoring,
        stop_performance_monitoring,
        get_performance_report,
        check_performance_alerts,
        start_performance_dashboard
    )

    if args.start:
        print(f"📊 Starting performance monitoring (interval: {args.interval}s)...")
        start_performance_monitoring(args.interval)
        print("✅ Performance monitoring started. Press Ctrl+C to stop.")
        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 Stopping performance monitoring...")
            stop_performance_monitoring()
            print("✅ Performance monitoring stopped.")

    elif args.stop:
        print("🛑 Stopping performance monitoring...")
        stop_performance_monitoring()
        print("✅ Performance monitoring stopped.")

    elif args.report:
        fmt = "json" if as_json else "text"
        report = get_performance_report(fmt)
        print(report)

    elif args.dashboard:
        print(f"🚀 Starting performance dashboard at http://{args.host}:{args.port}")
        start_performance_dashboard(args.host, args.port)
        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 Stopping dashboard...")

    elif args.alerts:
        alerts = check_performance_alerts()
        if as_json:
            _print_json({"alerts": alerts})
            return 0
        if not alerts:
            print("✅ No performance alerts found.")
        else:
            print(f"⚠️  Found {len(alerts)} performance alert(s):")
            for alert in alerts:
                icon = "🔴" if alert["severity"] == "high" else "🟡" if alert["severity"] == "medium" else "🟢"
                print(f"  {icon} {alert['type'].upper()}: {alert['message']}")

    else:
        print("Use --start, --stop, --report, --dashboard, or --alerts")
        return 1

    return 0


# ─────────────────────────────────────────────────────────────────────────────
# completion command
# ─────────────────────────────────────────────────────────────────────────────

_BASH_COMPLETION = """\
# flypy bash completion
# Add to ~/.bashrc: source <(flypy completion bash)
_flypy_completion() {
    local cur prev words cword
    _init_completion || return

    local commands="init build deploy list local verify inspect clean monitor completion"
    local global_opts="--version --config --help"

    case "${prev}" in
        flypy)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return ;;
        init)
            COMPREPLY=($(compgen -W "--output --template --force" -- "${cur}"))
            return ;;
        build)
            COMPREPLY=($(compgen -W "--output --mode --verbose --quiet --json --go-binary --optimize --no-optimize --optimization-level --no-cold-start-optimize --no-parallel --no-incremental --max-workers" -- "${cur}"))
            return ;;
        deploy)
            COMPREPLY=($(compgen -W "--registry --token --app-id --provider --region --quiet --json" -- "${cur}"))
            return ;;
        list)
            COMPREPLY=($(compgen -W "--json --quiet" -- "${cur}"))
            return ;;
        local)
            COMPREPLY=($(compgen -W "--port --watch --quiet" -- "${cur}"))
            return ;;
        verify)
            COMPREPLY=($(compgen -W "--json --quiet" -- "${cur}"))
            return ;;
        inspect)
            COMPREPLY=($(compgen -W "--json" -- "${cur}"))
            return ;;
        clean)
            COMPREPLY=($(compgen -W "--output --cache --dry-run --quiet" -- "${cur}"))
            return ;;
        monitor)
            COMPREPLY=($(compgen -W "--start --stop --report --dashboard --alerts --interval --host --port --json" -- "${cur}"))
            return ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return ;;
        --template)
            COMPREPLY=($(compgen -W "basic calculator data-transform api-call" -- "${cur}"))
            return ;;
        --mode)
            COMPREPLY=($(compgen -W "deterministic compatible" -- "${cur}"))
            return ;;
        --optimization-level)
            COMPREPLY=($(compgen -W "minimal balanced aggressive" -- "${cur}"))
            return ;;
        --provider)
            COMPREPLY=($(compgen -W "cloudflare vercel fly deno" -- "${cur}"))
            return ;;
    esac

    COMPREPLY=($(compgen -W "${global_opts} ${commands}" -- "${cur}"))
}
complete -F _flypy_completion flypy
"""

_ZSH_COMPLETION = """\
#compdef flypy
# flypy zsh completion
# Add to ~/.zshrc: source <(flypy completion zsh)

_flypy() {
    local state

    _arguments \\
        '--version[Show version]' \\
        '--config[Config file path]:file:_files' \\
        '1: :->command' \\
        '*: :->args'

    case $state in
        command)
            local commands=(
                'init:Scaffold a new FlyPy function file'
                'build:Build FlyPy functions to WebAssembly'
                'deploy:Deploy FlyPy functions to FunctionFly'
                'list:List registered FlyPy functions'
                'local:Run FlyPy functions locally for testing'
                'verify:Verify determinism of built artifacts'
                'inspect:Show detailed information about a built artifact'
                'clean:Remove build artifacts and cache'
                'monitor:Performance monitoring and profiling'
                'completion:Generate shell completion scripts'
            )
            _describe 'command' commands ;;
        args)
            case $words[2] in
                build)
                    _arguments \\
                        '--output[Output directory]:dir:_files -/' \\
                        '--mode[Execution mode]:(deterministic compatible)' \\
                        '--verbose[Verbose output]' \\
                        '--quiet[Quiet mode]' \\
                        '--json[JSON output]' \\
                        '--optimize[Enable optimization]' \\
                        '--no-optimize[Disable optimization]' \\
                        '--optimization-level[Optimization level]:(minimal balanced aggressive)' ;;
                completion)
                    _arguments '1: :(bash zsh fish)' ;;
            esac ;;
    esac
}

_flypy
"""

_FISH_COMPLETION = """\
# flypy fish completion
# Save to ~/.config/fish/completions/flypy.fish

set -l commands init build deploy list local verify inspect clean monitor completion

complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a init -d "Scaffold a new FlyPy function file"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a build -d "Build FlyPy functions to WebAssembly"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a deploy -d "Deploy FlyPy functions to FunctionFly"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a list -d "List registered FlyPy functions"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a local -d "Run FlyPy functions locally"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a verify -d "Verify determinism of built artifacts"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a inspect -d "Show artifact details"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a clean -d "Remove build artifacts"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a monitor -d "Performance monitoring"
complete -c flypy -f -n "not __fish_seen_subcommand_from $commands" -a completion -d "Generate shell completion"

# build options
complete -c flypy -n "__fish_seen_subcommand_from build" -l output -s o -d "Output directory"
complete -c flypy -n "__fish_seen_subcommand_from build" -l mode -d "Execution mode" -a "deterministic compatible"
complete -c flypy -n "__fish_seen_subcommand_from build" -l verbose -s v -d "Verbose output"
complete -c flypy -n "__fish_seen_subcommand_from build" -l quiet -s q -d "Quiet mode"
complete -c flypy -n "__fish_seen_subcommand_from build" -l json -d "JSON output"
complete -c flypy -n "__fish_seen_subcommand_from build" -l optimize -d "Enable optimization"
complete -c flypy -n "__fish_seen_subcommand_from build" -l no-optimize -d "Disable optimization"
complete -c flypy -n "__fish_seen_subcommand_from build" -l optimization-level -d "Optimization level" -a "minimal balanced aggressive"

# init options
complete -c flypy -n "__fish_seen_subcommand_from init" -l template -s t -d "Template" -a "basic calculator data-transform api-call"
complete -c flypy -n "__fish_seen_subcommand_from init" -l output -s o -d "Output directory"
complete -c flypy -n "__fish_seen_subcommand_from init" -l force -s f -d "Overwrite existing files"

# completion shells
complete -c flypy -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
"""


def completion_command(args) -> int:
    """Generate shell completion scripts."""
    shell = args.shell
    if shell == "bash":
        print(_BASH_COMPLETION)
    elif shell == "zsh":
        print(_ZSH_COMPLETION)
    elif shell == "fish":
        print(_FISH_COMPLETION)
    else:
        print(f"Unknown shell: {shell}", file=sys.stderr)
        return 1

    if shell == "bash":
        print("# Usage: source <(flypy completion bash)", file=sys.stderr)
    elif shell == "zsh":
        print("# Usage: source <(flypy completion zsh)", file=sys.stderr)
    elif shell == "fish":
        print("# Usage: flypy completion fish > ~/.config/fish/completions/flypy.fish", file=sys.stderr)

    return 0


# ─────────────────────────────────────────────────────────────────────────────
# Internal helpers
# ─────────────────────────────────────────────────────────────────────────────

def _load_python_file(file_path: str) -> Any:
    """Load a Python file to register FlyPy functions and return the module."""
    try:
        spec = importlib.util.spec_from_file_location("flypy_module", file_path)
        if spec and spec.loader:
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            return module
    except Exception as e:
        print(f"Warning: Could not load {file_path}: {e}", file=sys.stderr)
        return None


def _get_function_from_module(module: Any, function_name: str) -> Optional[Callable]:
    """Get the actual callable function from a loaded module by FlyPy name."""
    if not module:
        return None

    # Direct name match
    func = getattr(module, function_name, None)
    if func and callable(func) and hasattr(func, "_flypy_metadata"):
        return func

    # Hyphen → underscore
    python_name = function_name.replace("-", "_")
    func = getattr(module, python_name, None)
    if func and callable(func) and hasattr(func, "_flypy_metadata"):
        return func

    # Search all attributes
    for attr_name in dir(module):
        if not attr_name.startswith("_"):
            attr = getattr(module, attr_name)
            if callable(attr) and hasattr(attr, "_flypy_metadata"):
                if attr._flypy_metadata.name == function_name:
                    return attr

    return None


def _build_function(
    go_binary: Optional[str],
    func_def,
    output_dir: Path,
    mode: str,
    verbose: bool,
    optimize_bundle: bool = True,
    optimization_level: str = "balanced",
    optimize_cold_start: bool = True
) -> BuildResult:
    """Build a single function using the Go compiler."""
    import tempfile

    func_name = func_def.metadata.name
    start_time = time.time()

    if not go_binary:
        return BuildResult(
            success=False,
            function_name=func_name,
            output_dir=str(output_dir / func_name),
            errors=[
                "FlyPy Go binary not found",
                "Install: curl -sSL https://get.functionfly.com | sh",
                "Or specify: flypy build handler.py --go-binary /path/to/flypy-go",
            ],
        )

    with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
        f.write(func_def.source_code)
        temp_file = f.name

    try:
        cmd = [
            go_binary,
            "compile",
            "--input", temp_file,
            "--output", str(output_dir / func_name),
            "--mode", mode,
        ]

        if optimize_bundle:
            cmd.extend(["--optimize", optimization_level])
        else:
            cmd.append("--no-optimize")

        if optimize_cold_start:
            cmd.append("--optimize-cold-start")
        else:
            cmd.append("--no-cold-start-optimization")

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
            func_output_dir = output_dir / func_name
            wasm_file = func_output_dir / "state_transition.wasm"
            manifest_file = func_output_dir / "manifest.json"

            wasm_size = wasm_file.stat().st_size if wasm_file.exists() else None

            return BuildResult(
                success=True,
                function_name=func_name,
                output_dir=str(func_output_dir),
                wasm_file=str(wasm_file) if wasm_file.exists() else None,
                manifest_file=str(manifest_file) if manifest_file.exists() else None,
                build_time_ms=build_time,
                wasm_size_bytes=wasm_size,
            )
        else:
            errors = []
            if result.stderr:
                errors.extend(result.stderr.strip().split("\n"))
            if result.stdout and verbose:
                print(f"Go compiler stdout: {result.stdout}")

            return BuildResult(
                success=False,
                function_name=func_name,
                output_dir=str(output_dir / func_name),
                errors=errors,
                build_time_ms=build_time,
            )

    finally:
        try:
            os.unlink(temp_file)
        except Exception:
            pass


def _find_go_binary() -> Optional[str]:
    """Find the FlyPy Go binary in PATH or common locations."""
    # Check PATH via shutil.which
    found = shutil.which("flypy-go")
    if found:
        return found

    # Check ~/bin/flypy-go
    home_bin = Path.home() / "bin" / "flypy-go"
    if home_bin.exists() and home_bin.is_file():
        return str(home_bin)

    # Check relative to this script
    script_dir = Path(__file__).parent.parent.parent.parent
    candidates = [
        script_dir / "cmd" / "flypy-go" / "flypy-go",
        script_dir / "cmd" / "flypy" / "flypy",
        script_dir / "bin" / "flypy-go",
        script_dir / "flypy-go",
    ]

    for candidate in candidates:
        if candidate.exists() and candidate.is_file():
            return str(candidate)

    return None


if __name__ == "__main__":
    sys.exit(main())
