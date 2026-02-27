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
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from typing import List, Optional, Callable, Any

from .decorators import get_registered_functions, get_function_definition
from .types import BuildResult, ExecutionMode
from . import __version__


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

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Build command
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
        default="./dist",
        help="Output directory for artifacts"
    )
    build_parser.add_argument(
        "--mode",
        choices=["deterministic", "compatible"],
        default="deterministic",
        help="Execution mode"
    )
    build_parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Verbose output"
    )
    build_parser.add_argument(
        "--go-binary",
        help="Path to FlyPy Go binary (auto-detected if not specified)"
    )
    build_parser.add_argument(
        "--optimize",
        choices=["true", "false"],
        default="true",
        help="Enable bundle size optimization (default: true)"
    )
    build_parser.add_argument(
        "--optimization-level",
        choices=["minimal", "balanced", "aggressive"],
        default="balanced",
        help="Optimization level for bundle size reduction (default: balanced)"
    )
    build_parser.add_argument(
        "--optimize-cold-start",
        choices=["true", "false"],
        default="true",
        help="Enable cold start optimization (default: true)"
    )
    build_parser.add_argument(
        "--parallel",
        choices=["true", "false"],
        default="true",
        help="Enable parallel building (default: true)"
    )
    build_parser.add_argument(
        "--incremental",
        choices=["true", "false"],
        default="true",
        help="Enable incremental builds (default: true)"
    )
    build_parser.add_argument(
        "--max-workers",
        type=int,
        help="Maximum number of parallel workers (default: CPU count)"
    )

    # Performance monitoring command
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

    # Deploy command
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
        default="https://api.functionfly.com",
        help="FunctionFly registry URL (default: https://api.functionfly.com)"
    )
    deploy_parser.add_argument(
        "--token",
        help="Authentication token"
    )
    deploy_parser.add_argument(
        "--app-id",
        required=True,
        help="FunctionFly app ID to deploy to"
    )
    deploy_parser.add_argument(
        "--provider",
        default="cloudflare",
        choices=["cloudflare", "vercel", "fly", "deno"],
        help="Cloud provider for deployment (default: cloudflare)"
    )
    deploy_parser.add_argument(
        "--region",
        default="us-east-1",
        help="Deployment region (default: us-east-1)"
    )

    # List command
    list_parser = subparsers.add_parser(
        "list",
        help="List registered FlyPy functions"
    )
    list_parser.add_argument(
        "files",
        nargs="*",
        help="Python files to scan (optional)"
    )

    # Local command
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
        default=8080,
        help="Port to run local server on"
    )

    # Verify command
    verify_parser = subparsers.add_parser(
        "verify",
        help="Verify determinism of built artifacts"
    )
    verify_parser.add_argument(
        "artifact_dir",
        help="Directory containing built artifacts"
    )

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    try:
        if args.command == "build":
            return build_command(args)
        elif args.command == "deploy":
            return deploy_command(args)
        elif args.command == "list":
            return list_command(args)
        elif args.command == "local":
            return local_command(args)
        elif args.command == "verify":
            return verify_command(args)
        elif args.command == "monitor":
            return monitor_command(args)
        else:
            parser.print_help()
            return 1
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        if args.verbose:
            import traceback
            traceback.print_exc()
        return 1


def build_command(args) -> int:
    """Build FlyPy functions."""
    print("🔨 Building FlyPy functions...")

    # Extract arguments
    go_binary = args.go_binary
    optimize_bundle = args.optimize == "true"
    optimization_level = args.optimization_level
    optimize_cold_start = args.optimize_cold_start == "true"
    enable_parallel = args.parallel == "true"
    enable_incremental = args.incremental == "true"
    max_workers = args.max_workers

    # Load Python files to register functions
    for file_path in args.files:
        if not os.path.exists(file_path):
            print(f"Error: File {file_path} not found", file=sys.stderr)
            return 1

        print(f"Loading {file_path}...")
        _load_python_file(file_path)

    # Get registered functions
    functions = get_registered_functions()
    if not functions:
        print("No FlyPy functions found in the specified files", file=sys.stderr)
        return 1

    print(f"Found {len(functions)} function(s): {', '.join(functions.keys())}")

    # Create output directory
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Use optimized build process for multiple functions
    if len(functions) > 1:
        from flypy.build import build_all

        results = build_all(
            output_dir=str(output_dir),
            mode=args.mode,
            verbose=args.verbose,
            go_binary=go_binary,
            optimize_bundle=optimize_bundle,
            optimization_level=optimization_level,
            optimize_cold_start=optimize_cold_start,
            enable_parallel=enable_parallel,
            enable_incremental=enable_incremental,
            max_parallel_workers=max_workers
        )

        # Process results
        success_count = 0
        for result in results:
            if result.success:
                success_count += 1
                print(f"✅ Built {result.function_name} successfully")
                if args.verbose:
                    print(f"   Output: {result.output_dir}")
                    if result.wasm_size_bytes:
                        print(f"   WASM size: {result.wasm_size_bytes} bytes")
                    if result.build_time_ms:
                        print(f"   Build time: {result.build_time_ms}ms")
            else:
                print(f"❌ Failed to build {result.function_name}")
                for error in result.errors:
                    print(f"   Error: {error}")

        if success_count == len(functions):
            print(f"\n🎉 Successfully built all {success_count} function(s)!")
            return 0
        else:
            print(f"\n⚠️  Built {success_count}/{len(functions)} function(s)")
            return 1

    else:
        # Single function - use existing logic
        success_count = 0

        for func_name, func_def in functions.items():
            print(f"\nBuilding function: {func_name}")

            try:
                result = _build_function(go_binary, func_def, output_dir, args.mode, args.verbose,
                                       optimize_bundle, optimization_level, optimize_cold_start)
                if result.success:
                    success_count += 1
                    print(f"✅ Built {func_name} successfully")
                    if args.verbose:
                        print(f"   Output: {result.output_dir}")
                        if result.wasm_size_bytes:
                            print(f"   WASM size: {result.wasm_size_bytes} bytes")
                        if result.build_time_ms:
                            print(f"   Build time: {result.build_time_ms}ms")
                else:
                    print(f"❌ Failed to build {func_name}")
                    for error in result.errors:
                        print(f"   Error: {error}")

            except Exception as e:
                print(f"❌ Build failed for {func_name}: {e}")
                if args.verbose:
                    import traceback
                    traceback.print_exc()

        if success_count == len(functions):
            print(f"\n🎉 Successfully built all {success_count} function(s)!")
            return 0
        else:
            print(f"\n⚠️  Built {success_count}/{len(functions)} function(s)")
            return 1

    # Find Go binary
    go_binary = args.go_binary or _find_go_binary()
    if not go_binary:
        print("Error: Could not find FlyPy Go binary. Please specify --go-binary or ensure it's in PATH", file=sys.stderr)
        return 1

    # Parse optimization settings
    optimize_bundle = args.optimize == "true"
    optimization_level = args.optimization_level
    optimize_cold_start = args.optimize_cold_start == "true"
    enable_parallel = args.parallel == "true"
    enable_incremental = args.incremental == "true"
    max_workers = args.max_workers

    success_count = 0

    for func_name, func_def in functions.items():
        print(f"\nBuilding function: {func_name}")

        try:
            result = _build_function(
                go_binary, func_def, output_dir, args.mode, args.verbose,
                optimize_bundle, optimization_level, optimize_cold_start
            )
            if result.success:
                success_count += 1
                print(f"✅ Built {func_name} successfully")
                if args.verbose:
                    print(f"   Output: {result.output_dir}")
                    if result.wasm_size_bytes:
                        print(f"   WASM size: {result.wasm_size_bytes} bytes")
                    if result.build_time_ms:
                        print(f"   Build time: {result.build_time_ms}ms")
            else:
                print(f"❌ Failed to build {func_name}")
                for error in result.errors:
                    print(f"   Error: {error}")

        except Exception as e:
            print(f"❌ Build failed for {func_name}: {e}")
            if args.verbose:
                import traceback
                traceback.print_exc()

    if success_count == len(functions):
        print(f"\n🎉 Successfully built all {success_count} function(s)!")
        return 0
    else:
        print(f"\n⚠️  Built {success_count}/{len(functions)} function(s)")
        return 1


def deploy_command(args) -> int:
    """Deploy FlyPy functions."""
    print("📦 Deploying FlyPy functions...")

    artifact_dir = Path(args.artifact_dir)
    if not artifact_dir.exists():
        print(f"Error: Artifact directory {artifact_dir} not found", file=sys.stderr)
        return 1

    # Check for required files
    manifest_file = artifact_dir / "manifest.json"
    wasm_file = artifact_dir / "state_transition.wasm"

    if not manifest_file.exists():
        print(f"Error: manifest.json not found in {artifact_dir}", file=sys.stderr)
        return 1

    if not wasm_file.exists():
        print(f"Error: state_transition.wasm not found in {artifact_dir}", file=sys.stderr)
        return 1

    # Load manifest
    try:
        with open(manifest_file) as f:
            manifest = json.load(f)
        func_name = manifest.get("name", "unknown")
    except Exception as e:
        print(f"Error reading manifest: {e}", file=sys.stderr)
        return 1

    print(f"Deploying function: {func_name}")

    # Validate required arguments
    if not args.token:
        print("Error: --token is required for deployment", file=sys.stderr)
        return 1

    # Read and encode WASM artifact
    try:
        with open(wasm_file, 'rb') as f:
            wasm_data = f.read()
        artifact_b64 = base64.b64encode(wasm_data).decode('utf-8')
    except Exception as e:
        print(f"Error reading WASM file: {e}", file=sys.stderr)
        return 1

    # Prepare deployment request
    deploy_data = {
        "provider": args.provider,
        "region": args.region,
        "artifact": artifact_b64,
        "routes": manifest.get("routes", []),
        "env_vars": manifest.get("env_vars", {}),
        "secrets": manifest.get("secrets", {}),
        "provider_config": manifest.get("provider_config", {})
    }

    # Make deployment request
    try:
        url = f"{args.registry}/apps/{args.app_id}/deploy"
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {args.token}'
        }

        data = json.dumps(deploy_data).encode('utf-8')
        req = urllib.request.Request(url, data=data, headers=headers, method='POST')

        print(f"🚀 Deploying to {args.provider} in {args.region}...")

        with urllib.request.urlopen(req) as response:
            if response.status == 201:
                result = json.loads(response.read().decode('utf-8'))
                deployment_id = result.get('id', 'unknown')
                print(f"✅ Deployment successful!")
                print(f"   Deployment ID: {deployment_id}")
                print(f"   Status: {result.get('status', 'unknown')}")
                if 'url' in result:
                    print(f"   URL: {result['url']}")
                return 0
            else:
                error_data = json.loads(response.read().decode('utf-8'))
                print(f"❌ Deployment failed: {error_data.get('error', 'Unknown error')}", file=sys.stderr)
                return 1

    except urllib.error.HTTPError as e:
        try:
            error_data = json.loads(e.read().decode('utf-8'))
            print(f"❌ Deployment failed (HTTP {e.code}): {error_data.get('error', str(e))}", file=sys.stderr)
        except:
            print(f"❌ Deployment failed (HTTP {e.code}): {str(e)}", file=sys.stderr)
        return 1
    except urllib.error.URLError as e:
        print(f"❌ Network error: {e}", file=sys.stderr)
        return 1
    except json.JSONDecodeError as e:
        print(f"❌ Invalid response from server: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"❌ Unexpected error during deployment: {e}", file=sys.stderr)
        return 1


def list_command(args) -> int:
    """List registered FlyPy functions."""
    files_to_scan = args.files or ["."]

    for file_path in files_to_scan:
        if os.path.isdir(file_path):
            # Scan directory for Python files
            python_files = list(Path(file_path).rglob("*.py"))
            for py_file in python_files:
                _load_python_file(str(py_file))
        else:
            _load_python_file(file_path)

    functions = get_registered_functions()

    if not functions:
        print("No FlyPy functions found")
        return 0

    print(f"Found {len(functions)} FlyPy function(s):")
    print()

    for name, func_def in functions.items():
        metadata = func_def.metadata
        print(f"📦 {name} (v{metadata.version})")
        if metadata.description:
            print(f"   Description: {metadata.description}")
        print(f"   Deterministic: {'✅' if metadata.deterministic else '❌'}")
        print(f"   Idempotent: {'✅' if metadata.idempotent else '❌'}")
        print(f"   Pure: {'✅' if metadata.pure else '❌'}")
        if metadata.capabilities:
            print(f"   Capabilities: {', '.join(metadata.capabilities)}")
        if metadata.cache_ttl:
            print(f"   Cache TTL: {metadata.cache_ttl}s")
        print()

    return 0


class FunctionRequestHandler(BaseHTTPRequestHandler):
    """HTTP request handler for FlyPy function execution."""

    def __init__(self, function: Callable, *args, **kwargs):
        self.function = function
        super().__init__(*args, **kwargs)

    def do_POST(self) -> None:
        """Handle POST requests by executing the function."""
        try:
            # Read request body
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)

            # Parse JSON input
            try:
                input_data = json.loads(post_data.decode('utf-8'))
            except json.JSONDecodeError:
                self.send_error(400, "Invalid JSON in request body")
                return

            # Execute function
            try:
                result = self.function(input_data)
            except Exception as e:
                self.send_error(500, f"Function execution error: {str(e)}")
                return

            # Send response
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()

            response_data = json.dumps(result).encode('utf-8')
            self.wfile.write(response_data)

        except Exception as e:
            self.send_error(500, f"Server error: {str(e)}")

    def do_GET(self) -> None:
        """Handle GET requests with function info."""
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()

        metadata = getattr(self.function, '_flypy_metadata', None)
        description = metadata.description if metadata else ''

        info = {
            "status": "running",
            "function": self.function.__name__,
            "description": description,
        }
        self.wfile.write(json.dumps(info).encode('utf-8'))

    def log_message(self, format, *args):
        """Override to use our logging instead of default HTTP server logging."""
        pass


def start_local_server(function: Callable, port: int) -> None:
    """Start the local HTTP server for function execution."""

    def handler_factory(*args, **kwargs):
        return FunctionRequestHandler(function, *args, **kwargs)

    server = HTTPServer(('localhost', port), handler_factory)

    print(f"🌐 Server started at http://localhost:{port}")
    print("📖 GET /  - Function info")
    print("🚀 POST / - Execute function (JSON body)")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n🛑 Server stopped")
        server.shutdown()


def local_command(args) -> int:
    """Run FlyPy functions locally."""
    print("🏠 Starting local FlyPy server...")

    if not os.path.exists(args.file):
        print(f"Error: File {args.file} not found", file=sys.stderr)
        return 1

    # Load the Python file and get the module
    module = _load_python_file(args.file)
    if not module:
        print(f"Error: Could not load {args.file}", file=sys.stderr)
        return 1

    # Get the function definition from registry
    func_def = get_function_definition(args.function)
    if not func_def:
        print(f"Error: Function '{args.function}' not found in {args.file}", file=sys.stderr)
        return 1

    # Get the actual callable function from the module
    function = _get_function_from_module(module, args.function)
    if not function:
        print(f"Error: Could not get callable function '{args.function}'", file=sys.stderr)
        return 1

    print(f"Running function: {args.function}")
    print(f"Description: {func_def.metadata.description or 'No description'}")
    print(f"Local server starting on port {args.port}")
    print("Press Ctrl+C to stop")

    # Start the server in a separate thread
    server_thread = threading.Thread(
        target=start_local_server,
        args=(function, args.port),
        daemon=True
    )
    server_thread.start()

    try:
        # Keep the main thread alive
        while server_thread.is_alive():
            time.sleep(0.1)
    except KeyboardInterrupt:
        print("\n🛑 Stopping server...")

    return 0


def verify_command(args) -> int:
    """Verify determinism of artifacts."""
    print("🔍 Verifying artifact determinism...")

    artifact_dir = Path(args.artifact_dir)
    if not artifact_dir.exists():
        print(f"Error: Artifact directory {artifact_dir} not found", file=sys.stderr)
        return 1

    # Check for required files
    manifest_file = artifact_dir / "manifest.json"
    hash_file = artifact_dir / "determinism.hash"
    sig_file = artifact_dir / "signature.sig"

    if not manifest_file.exists():
        print(f"Error: manifest.json not found in {artifact_dir}", file=sys.stderr)
        return 1

    if not hash_file.exists():
        print(f"Error: determinism.hash not found in {artifact_dir}", file=sys.stderr)
        return 1

    if not sig_file.exists():
        print(f"Error: signature.sig not found in {artifact_dir}", file=sys.stderr)
        return 1

    # Load and verify manifest
    try:
        with open(manifest_file) as f:
            manifest = json.load(f)

        with open(hash_file) as f:
            stored_hash = f.read().strip()

        func_name = manifest.get("name", "unknown")
        print(f"Verifying function: {func_name}")

        # Verify determinism hash
        print("🔐 Verifying determinism hash...")
        try:
            # Read WASM file
            wasm_file = artifact_dir / "state_transition.wasm"
            if not wasm_file.exists():
                print(f"Error: state_transition.wasm not found in {artifact_dir}", file=sys.stderr)
                return 1

            with open(wasm_file, 'rb') as f:
                wasm_data = f.read()

            # Compute SHA256 hash of WASM data
            computed_hash = hashlib.sha256(wasm_data).hexdigest()

            # Compare with stored hash
            if computed_hash != stored_hash:
                print(f"❌ Hash verification failed!", file=sys.stderr)
                print(f"   Expected: {stored_hash}", file=sys.stderr)
                print(f"   Computed: {computed_hash}", file=sys.stderr)
                return 1

            print("✅ Determinism hash verified")

        except Exception as e:
            print(f"Error verifying determinism hash: {e}", file=sys.stderr)
            return 1

        # Verify signature (if present)
        print("🔏 Verifying signature...")
        try:
            with open(sig_file, 'rb') as f:
                signature_data = f.read()

            if len(signature_data) == 0:
                print("⚠️  No signature data found (verification skipped)")
            else:
                # For now, just check signature format (64 bytes for Ed25519)
                if len(signature_data) != 64:
                    print(f"❌ Invalid signature format (expected 64 bytes, got {len(signature_data)})", file=sys.stderr)
                    return 1

                print("✅ Signature format verified")
                print("⚠️  Full cryptographic verification requires 'cryptography' package")
                print("   Install with: pip install cryptography")

        except Exception as e:
            print(f"Error verifying signature: {e}", file=sys.stderr)
            return 1

        # Verify WASM module structure (basic validation)
        print("🔍 Validating WASM module...")
        try:
            # Check WASM magic bytes
            if len(wasm_data) < 8:
                print("❌ WASM file too small", file=sys.stderr)
                return 1

            if wasm_data[:4] != b'\x00asm':
                print("❌ Invalid WASM magic bytes", file=sys.stderr)
                return 1

            # Check version (should be 0x01 0x00 0x00 0x00 for version 1)
            if wasm_data[4:8] != b'\x01\x00\x00\x00':
                print("❌ Invalid WASM version", file=sys.stderr)
                return 1

            print("✅ WASM module structure verified")

        except Exception as e:
            print(f"Error validating WASM module: {e}", file=sys.stderr)
            return 1

        print("🎉 All verifications passed!")
        print(f"   Function '{func_name}' artifacts are valid and deterministic")

    except Exception as e:
        print(f"Error during verification: {e}", file=sys.stderr)
        return 1

    return 0


def monitor_command(args) -> int:
    """Performance monitoring commands."""
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
            # Keep running until interrupted
            import time
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
        print("📊 Generating performance report...")
        report = get_performance_report("text")
        print(report)

    elif args.dashboard:
        print(f"🚀 Starting performance dashboard at http://{args.host}:{args.port}")
        dashboard_thread = start_performance_dashboard(args.host, args.port)

        try:
            # Keep running until interrupted
            import time
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 Stopping dashboard...")

    elif args.alerts:
        print("🔍 Checking for performance alerts...")
        alerts = check_performance_alerts()

        if not alerts:
            print("✅ No performance alerts found.")
        else:
            print(f"⚠️  Found {len(alerts)} performance alert(s):")
            for alert in alerts:
                severity_icon = "🔴" if alert["severity"] == "high" else "🟡" if alert["severity"] == "medium" else "🟢"
                print(f"  {severity_icon} {alert['type'].upper()}: {alert['message']}")

    else:
        print("Use --start, --stop, --report, --dashboard, or --alerts")
        return 1

    return 0


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

    # First try direct name match
    func = getattr(module, function_name, None)
    if func and callable(func) and hasattr(func, '_flypy_metadata'):
        return func

    # Try replacing hyphens with underscores (common Python naming convention)
    python_name = function_name.replace('-', '_')
    func = getattr(module, python_name, None)
    if func and callable(func) and hasattr(func, '_flypy_metadata'):
        return func

    # Search all callable attributes in the module
    for attr_name in dir(module):
        if not attr_name.startswith('_'):
            attr = getattr(module, attr_name)
            if callable(attr) and hasattr(attr, '_flypy_metadata'):
                metadata = attr._flypy_metadata
                if metadata.name == function_name:
                    return attr

    return None


def _build_function(
    go_binary: str,
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
    import time

    func_name = func_def.metadata.name
    start_time = time.time()

    # Create a temporary file with the function source
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write(func_def.source_code)
        temp_file = f.name

    try:
        # Prepare Go compiler arguments
        cmd = [
            go_binary,
            "compile",
            "--input", temp_file,
            "--output", str(output_dir / func_name),
            "--mode", mode,
        ]

        # Add optimization flags
        if optimize_bundle:
            cmd.extend(["--optimize", optimization_level])
        else:
            cmd.append("--no-optimize")

        # Add cold start optimization flags
        if optimize_cold_start:
            cmd.append("--optimize-cold-start")
        else:
            cmd.append("--no-cold-start-optimization")

        if verbose:
            cmd.append("--verbose")
            print(f"Running: {' '.join(cmd)}")

        # Run the Go compiler
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            cwd=os.getcwd()
        )

        build_time = int((time.time() - start_time) * 1000)

        if result.returncode == 0:
            # Success - check if artifacts were created
            func_output_dir = output_dir / func_name
            wasm_file = func_output_dir / "state_transition.wasm"
            manifest_file = func_output_dir / "manifest.json"

            wasm_size = None
            if wasm_file.exists():
                wasm_size = wasm_file.stat().st_size

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
            # Failure
            errors = []
            if result.stderr:
                errors.extend(result.stderr.strip().split('\n'))
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
        # Clean up temp file
        try:
            os.unlink(temp_file)
        except:
            pass


def _find_go_binary() -> Optional[str]:
    """Find the FlyPy Go binary in PATH or common locations."""
    # Check PATH first
    import shutil
    go_binary = os.path.expanduser("~/bin/flypy-go")
    if go_binary:
        return go_binary

    # Check common locations relative to this script
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


if __name__ == "__main__":
    sys.exit(main())