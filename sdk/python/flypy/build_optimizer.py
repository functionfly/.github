"""
Build-time optimizations for FlyPy functions.

This module implements parallel compilation, incremental builds, and build caching
to improve build performance and reduce build times.
"""

import hashlib
import json
import os
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Dict, Any, List, Optional, Set, Tuple
import subprocess


class BuildCache:
    """Cache for incremental builds."""

    def __init__(self, cache_dir: Optional[str] = None):
        """
        Initialize the build cache.

        Args:
            cache_dir: Directory to store build cache
        """
        self.cache_dir = Path(cache_dir) if cache_dir else Path.home() / ".flypy" / "build_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.cache_index_file = self.cache_dir / "cache_index.json"
        self._load_cache_index()

    def _load_cache_index(self):
        """Load the cache index from disk."""
        if self.cache_index_file.exists():
            try:
                with open(self.cache_index_file, 'r') as f:
                    self._cache_index = json.load(f)
            except Exception:
                self._cache_index = {}
        else:
            self._cache_index = {}

    def _save_cache_index(self):
        """Save the cache index to disk."""
        try:
            with open(self.cache_index_file, 'w') as f:
                json.dump(self._cache_index, f, indent=2)
        except Exception:
            pass

    def get_cached_build(self, function_name: str, source_hash: str, build_config: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """
        Get a cached build result.

        Args:
            function_name: Name of the function
            source_hash: Hash of the source code
            build_config: Build configuration

        Returns:
            Cached build result or None
        """
        cache_key = self._get_cache_key(function_name, source_hash, build_config)

        if cache_key not in self._cache_index:
            return None

        cache_info = self._cache_index[cache_key]
        cache_file = self.cache_dir / f"{cache_key}.json"

        # Check if cache file exists
        if not cache_file.exists():
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

        # Check if build config has changed
        if cache_info["build_config"] != build_config:
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

        try:
            with open(cache_file, 'r') as f:
                cached_result = json.load(f)

            # Verify cached artifacts still exist
            if cached_result.get("wasm_file") and not Path(cached_result["wasm_file"]).exists():
                del self._cache_index[cache_key]
                self._save_cache_index()
                return None

            return cached_result

        except Exception:
            # Cache corrupted, remove it
            if cache_file.exists():
                cache_file.unlink()
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

    def store_build_result(self, function_name: str, source_hash: str, build_config: Dict[str, Any], build_result: Dict[str, Any]):
        """
        Store a build result in cache.

        Args:
            function_name: Name of the function
            source_hash: Hash of the source code
            build_config: Build configuration
            build_result: Build result to cache
        """
        cache_key = self._get_cache_key(function_name, source_hash, build_config)
        cache_file = self.cache_dir / f"{cache_key}.json"

        # Clean up old cache entries if needed
        self._cleanup_old_entries()

        cache_info = {
            "function_name": function_name,
            "source_hash": source_hash,
            "build_config": build_config,
            "cached_at": time.time(),
            "wasm_size": build_result.get("wasm_size_bytes", 0)
        }

        try:
            # Store the cache info
            self._cache_index[cache_key] = cache_info
            self._save_cache_index()

            # Store the build result
            with open(cache_file, 'w') as f:
                json.dump(build_result, f, indent=2, default=str)

        except Exception:
            # If caching fails, just continue without caching
            pass

    def _get_cache_key(self, function_name: str, source_hash: str, build_config: Dict[str, Any]) -> str:
        """Generate a cache key."""
        config_str = json.dumps(build_config, sort_keys=True)
        combined = f"{function_name}:{source_hash}:{config_str}"
        return hashlib.md5(combined.encode()).hexdigest()

    def _cleanup_old_entries(self, max_entries: int = 100):
        """Clean up old cache entries."""
        if len(self._cache_index) <= max_entries:
            return

        # Sort by cache time and keep the most recent
        sorted_entries = sorted(
            self._cache_index.items(),
            key=lambda x: x[1]["cached_at"],
            reverse=True
        )

        # Remove old entries
        for cache_key, _ in sorted_entries[max_entries:]:
            cache_file = self.cache_dir / f"{cache_key}.json"
            if cache_file.exists():
                cache_file.unlink()
            del self._cache_index[cache_key]

        self._save_cache_index()


class ParallelBuilder:
    """Builds multiple functions in parallel."""

    def __init__(self, max_workers: Optional[int] = None):
        """
        Initialize the parallel builder.

        Args:
            max_workers: Maximum number of parallel workers
        """
        import multiprocessing
        self.max_workers = max_workers or min(multiprocessing.cpu_count(), 4)
        self.build_cache = BuildCache()

    def build_functions_parallel(
        self,
        functions: List[Dict[str, Any]],
        go_binary: str,
        output_dir: str,
        build_config: Dict[str, Any],
        verbose: bool = False
    ) -> List[Dict[str, Any]]:
        """
        Build multiple functions in parallel.

        Args:
            functions: List of function definitions to build
            go_binary: Path to Go binary
            output_dir: Output directory
            build_config: Build configuration
            verbose: Enable verbose output

        Returns:
            List of build results
        """
        results = []
        total_start_time = time.time()

        # Check cache for each function first
        cached_results = []
        functions_to_build = []

        for func_def in functions:
            func_name = func_def["metadata"]["name"]
            source_hash = func_def["metadata"]["source_hash"]

            cached_result = self.build_cache.get_cached_build(func_name, source_hash, build_config)
            if cached_result:
                cached_results.append(cached_result)
                if verbose:
                    print(f"✅ Using cached build for {func_name}")
            else:
                functions_to_build.append(func_def)

        # Build remaining functions in parallel
        if functions_to_build:
            if verbose:
                print(f"🔨 Building {len(functions_to_build)} functions in parallel (workers: {self.max_workers})")

            with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
                # Submit all build tasks
                future_to_func = {
                    executor.submit(self._build_single_function, func_def, go_binary, output_dir, build_config, verbose): func_def
                    for func_def in functions_to_build
                }

                # Collect results as they complete
                for future in as_completed(future_to_func):
                    func_def = future_to_func[future]
                    try:
                        result = future.result()
                        results.append(result)

                        # Cache successful builds
                        if result["success"]:
                            func_name = result["function_name"]
                            source_hash = func_def["metadata"]["source_hash"]
                            self.build_cache.store_build_result(func_name, source_hash, build_config, result)

                    except Exception as e:
                        func_name = func_def["metadata"]["name"]
                        error_result = {
                            "success": False,
                            "function_name": func_name,
                            "output_dir": str(Path(output_dir) / func_name),
                            "errors": [f"Build failed: {str(e)}"],
                            "build_time_ms": 0
                        }
                        results.append(error_result)

        # Combine cached and newly built results
        all_results = cached_results + results

        total_time = int((time.time() - total_start_time) * 1000)

        if verbose:
            cached_count = len(cached_results)
            built_count = len(results)
            print(f"📊 Build summary: {built_count} built, {cached_count} cached, total time: {total_time}ms")

        return all_results

    def _build_single_function(
        self,
        func_def: Dict[str, Any],
        go_binary: str,
        output_dir: str,
        build_config: Dict[str, Any],
        verbose: bool
    ) -> Dict[str, Any]:
        """Build a single function."""
        func_name = func_def["metadata"]["name"]
        start_time = time.time()

        try:
            # Create output directory
            output_path = Path(output_dir) / func_name
            output_path.mkdir(parents=True, exist_ok=True)

            # Create metadata file
            metadata_file = output_path / "function_metadata.json"
            with open(metadata_file, 'w') as f:
                json.dump(func_def, f, indent=2)

            # Create source file
            source_file = output_path / "function.py"
            with open(source_file, 'w') as f:
                f.write(func_def["source"]["code"])

            # Build command
            cmd = [
                go_binary,
                "compile",
                "--input", str(source_file),
                "--metadata", str(metadata_file),
                "--output", str(output_path),
                "--mode", build_config.get("mode", "deterministic"),
            ]

            # Add optimization flags
            if build_config.get("optimize_bundle", True):
                cmd.extend(["--optimize", build_config.get("optimization_level", "balanced")])
            else:
                cmd.append("--no-optimize")

            # Add cold start optimization flags
            if build_config.get("optimize_cold_start", True):
                cmd.append("--optimize-cold-start")
            else:
                cmd.append("--no-cold-start-optimization")

            # Filter out None values
            cmd = [arg for arg in cmd if arg is not None]

            if verbose:
                print(f"Running: {' '.join(cmd)}")

            # Execute build
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=os.getcwd(),
                timeout=300  # 5 minute timeout
            )

            build_time = int((time.time() - start_time) * 1000)

            if result.returncode == 0:
                # Check for artifacts
                wasm_file = output_path / "state_transition.wasm"
                manifest_file = output_path / "manifest.json"

                wasm_size = None
                if wasm_file.exists():
                    wasm_size = wasm_file.stat().st_size

                return {
                    "success": True,
                    "function_name": func_name,
                    "output_dir": str(output_path),
                    "wasm_file": str(wasm_file) if wasm_file.exists() else None,
                    "manifest_file": str(manifest_file) if manifest_file.exists() else None,
                    "build_time_ms": build_time,
                    "wasm_size_bytes": wasm_size,
                }
            else:
                errors = []
                if result.stderr:
                    errors.extend(result.stderr.strip().split('\n'))
                if not errors and result.stdout:
                    errors.extend(result.stdout.strip().split('\n'))

                return {
                    "success": False,
                    "function_name": func_name,
                    "output_dir": str(output_path),
                    "errors": errors,
                    "build_time_ms": build_time,
                }

        except subprocess.TimeoutExpired:
            build_time = int((time.time() - start_time) * 1000)
            return {
                "success": False,
                "function_name": func_name,
                "output_dir": str(Path(output_dir) / func_name),
                "errors": ["Build timeout after 5 minutes"],
                "build_time_ms": build_time,
            }

        except Exception as e:
            build_time = int((time.time() - start_time) * 1000)
            return {
                "success": False,
                "function_name": func_name,
                "output_dir": str(Path(output_dir) / func_name),
                "errors": [f"Build failed: {str(e)}"],
                "build_time_ms": build_time,
            }


class DependencyAnalyzer:
    """Analyzes function dependencies for optimized build ordering."""

    def __init__(self):
        self.dependency_graph: Dict[str, Set[str]] = {}

    def analyze_dependencies(self, functions: List[Dict[str, Any]]) -> List[List[Dict[str, Any]]]:
        """
        Analyze dependencies and return functions grouped by dependency level.

        Args:
            functions: List of function definitions

        Returns:
            Functions grouped by dependency level (can be built in parallel within levels)
        """
        # Build dependency graph
        func_by_name = {func["metadata"]["name"]: func for func in functions}

        for func in functions:
            func_name = func["metadata"]["name"]
            dependencies = set()

            # Check if function calls other functions
            # This is a simplified analysis - in practice, you'd do AST analysis
            source_code = func["source"]["code"]
            for other_func in functions:
                other_name = other_func["metadata"]["name"]
                if other_name != func_name and other_name in source_code:
                    dependencies.add(other_name)

            self.dependency_graph[func_name] = dependencies

        # Group functions by dependency level
        levels = []
        processed = set()

        while len(processed) < len(functions):
            current_level = []

            for func in functions:
                func_name = func["metadata"]["name"]
                if func_name in processed:
                    continue

                # Check if all dependencies are processed
                deps = self.dependency_graph[func_name]
                if all(dep in processed for dep in deps):
                    current_level.append(func)

            if not current_level:
                # Circular dependency or other issue
                # Add remaining functions to current level
                for func in functions:
                    func_name = func["metadata"]["name"]
                    if func_name not in processed:
                        current_level.append(func)
                break

            levels.append(current_level)
            processed.update(func["metadata"]["name"] for func in current_level)

        return levels

    def get_build_order(self, functions: List[Dict[str, Any]]) -> List[str]:
        """
        Get optimal build order considering dependencies.

        Args:
            functions: List of function definitions

        Returns:
            Function names in optimal build order
        """
        levels = self.analyze_dependencies(functions)
        build_order = []

        for level in levels:
            # Within each level, functions can be built in any order
            level_names = [func["metadata"]["name"] for func in level]
            build_order.extend(level_names)

        return build_order


class BuildOptimizer:
    """Main build optimizer coordinating all build-time optimizations."""

    def __init__(self):
        self.build_cache = BuildCache()
        self.parallel_builder = ParallelBuilder()
        self.dependency_analyzer = DependencyAnalyzer()
        self.build_stats = {
            "total_builds": 0,
            "cached_builds": 0,
            "parallel_builds": 0,
            "total_build_time_saved": 0
        }

    def optimize_build(
        self,
        functions: List[Dict[str, Any]],
        go_binary: str,
        output_dir: str,
        build_config: Dict[str, Any],
        enable_parallel: bool = True,
        enable_incremental: bool = True,
        verbose: bool = False
    ) -> Dict[str, Any]:
        """
        Optimize the build process for multiple functions.

        Args:
            functions: List of function definitions
            go_binary: Path to Go binary
            output_dir: Output directory
            build_config: Build configuration
            enable_parallel: Enable parallel building
            enable_incremental: Enable incremental builds
            verbose: Enable verbose output

        Returns:
            Build optimization results
        """
        start_time = time.time()

        # Analyze dependencies for optimal build ordering
        if enable_parallel:
            dependency_levels = self.dependency_analyzer.analyze_dependencies(functions)
            if verbose:
                print(f"📋 Dependency analysis: {len(dependency_levels)} levels identified")
        else:
            dependency_levels = [functions]  # Single level, sequential build

        all_results = []
        total_cached = 0
        total_built = 0

        # Build functions level by level
        for level_idx, level_functions in enumerate(dependency_levels):
            if verbose:
                print(f"🏗️  Building level {level_idx + 1}/{len(dependency_levels)} ({len(level_functions)} functions)")

            if enable_parallel and len(level_functions) > 1:
                # Build level in parallel
                level_results = self.parallel_builder.build_functions_parallel(
                    level_functions, go_binary, output_dir, build_config, verbose
                )
                total_built += len([r for r in level_results if not r.get("cached", False)])
                total_cached += len([r for r in level_results if r.get("cached", False)])
            else:
                # Build level sequentially
                level_results = []
                for func_def in level_functions:
                    result = self.parallel_builder._build_single_function(
                        func_def, go_binary, output_dir, build_config, verbose
                    )
                    level_results.append(result)
                    if result["success"]:
                        total_built += 1
                    else:
                        total_built += 1  # Count as attempted build

            all_results.extend(level_results)

        total_time = time.time() - start_time

        # Calculate statistics
        successful_builds = len([r for r in all_results if r["success"]])
        failed_builds = len(all_results) - successful_builds

        # Estimate time saved by caching and parallelization
        time_saved = total_cached * 2000  # Assume 2 seconds saved per cached build

        optimization_results = {
            "success": failed_builds == 0,
            "total_functions": len(functions),
            "successful_builds": successful_builds,
            "failed_builds": failed_builds,
            "cached_builds": total_cached,
            "parallel_enabled": enable_parallel,
            "incremental_enabled": enable_incremental,
            "total_build_time": total_time,
            "estimated_time_saved_ms": time_saved,
            "results": all_results,
            "dependency_levels": len(dependency_levels)
        }

        # Update global stats
        self.build_stats["total_builds"] += len(functions)
        self.build_stats["cached_builds"] += total_cached
        self.build_stats["parallel_builds"] += total_built if enable_parallel else 0
        self.build_stats["total_build_time_saved"] += time_saved

        return optimization_results

    def get_build_stats(self) -> Dict[str, Any]:
        """Get build optimization statistics."""
        return dict(self.build_stats)

    def clear_build_cache(self):
        """Clear the build cache."""
        # Remove all cache files
        import shutil
        if self.build_cache.cache_dir.exists():
            shutil.rmtree(self.build_cache.cache_dir)
        self.build_cache._cache_index = {}


# Convenience functions
def optimize_build_process(
    functions: List[Dict[str, Any]],
    go_binary: str,
    output_dir: str,
    build_config: Dict[str, Any],
    enable_parallel: bool = True,
    enable_incremental: bool = True,
    verbose: bool = False
) -> Dict[str, Any]:
    """
    Convenience function to optimize the build process.

    Args:
        functions: List of function definitions
        go_binary: Path to Go binary
        output_dir: Output directory
        build_config: Build configuration
        enable_parallel: Enable parallel building
        enable_incremental: Enable incremental builds
        verbose: Enable verbose output

    Returns:
        Build optimization results
    """
    optimizer = BuildOptimizer()
    return optimizer.optimize_build(
        functions, go_binary, output_dir, build_config,
        enable_parallel, enable_incremental, verbose
    )


def clear_build_cache():
    """Clear the global build cache."""
    optimizer = BuildOptimizer()
    optimizer.clear_build_cache()


def get_build_cache_stats() -> Dict[str, Any]:
    """Get build cache statistics."""
    optimizer = BuildOptimizer()
    return optimizer.get_build_stats()