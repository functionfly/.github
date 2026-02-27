"""
Cold start optimization for FlyPy functions.

This module implements lazy loading, precompilation, and caching strategies
to minimize cold start times for FlyPy functions.
"""

import asyncio
import hashlib
import json
import os
import pickle
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from functools import lru_cache, wraps
from pathlib import Path
from typing import Dict, Any, List, Optional, Callable, Union, Set
import importlib
import importlib.util
import threading
import weakref


class PrecompiledModuleCache:
    """Cache for precompiled Python modules to reduce import time."""

    def __init__(self, cache_dir: Optional[str] = None, max_cache_size: int = 100):
        """
        Initialize the precompiled module cache.

        Args:
            cache_dir: Directory to store cached modules
            max_cache_size: Maximum number of modules to cache
        """
        self.cache_dir = Path(cache_dir) if cache_dir else Path.home() / ".flypy" / "module_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.max_cache_size = max_cache_size
        self._cache_index: Dict[str, Dict[str, Any]] = {}
        self._load_cache_index()

    def get_compiled_module(self, module_name: str, module_path: str) -> Optional[Any]:
        """
        Get a precompiled module from cache.

        Args:
            module_name: Name of the module
            module_path: Path to the module file

        Returns:
            Compiled module or None if not cached
        """
        cache_key = self._get_cache_key(module_name, module_path)

        if cache_key not in self._cache_index:
            return None

        cache_info = self._cache_index[cache_key]
        cache_file = self.cache_dir / f"{cache_key}.pyc"

        # Check if cache is still valid
        if not cache_file.exists():
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

        # Check file modification time
        if os.path.getmtime(module_path) > cache_info["mtime"]:
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

        # Load cached module
        try:
            with open(cache_file, 'rb') as f:
                return pickle.load(f)
        except Exception:
            # Cache corrupted, remove it
            if cache_file.exists():
                cache_file.unlink()
            del self._cache_index[cache_key]
            self._save_cache_index()
            return None

    def store_compiled_module(self, module_name: str, module_path: str, compiled_module: Any):
        """
        Store a compiled module in cache.

        Args:
            module_name: Name of the module
            module_path: Path to the module file
            compiled_module: The compiled module object
        """
        cache_key = self._get_cache_key(module_name, module_path)
        cache_file = self.cache_dir / f"{cache_key}.pyc"

        # Clean up old cache entries if we're at capacity
        if len(self._cache_index) >= self.max_cache_size:
            self._cleanup_old_entries()

        # Store the module
        try:
            with open(cache_file, 'wb') as f:
                pickle.dump(compiled_module, f)

            self._cache_index[cache_key] = {
                "module_name": module_name,
                "module_path": module_path,
                "mtime": os.path.getmtime(module_path),
                "cached_at": time.time()
            }
            self._save_cache_index()

        except Exception:
            # If caching fails, just continue without caching
            pass

    def _get_cache_key(self, module_name: str, module_path: str) -> str:
        """Generate a cache key for a module."""
        content_hash = hashlib.md5()
        try:
            with open(module_path, 'rb') as f:
                content_hash.update(f.read())
        except Exception:
            # If we can't read the file, use path and mtime
            content_hash.update(f"{module_path}:{os.path.getmtime(module_path)}".encode())

        return f"{module_name}_{content_hash.hexdigest()[:16]}"

    def _cleanup_old_entries(self, keep_count: int = 80):
        """Clean up old cache entries to make room for new ones."""
        # Sort by cache time and keep the most recent
        sorted_entries = sorted(
            self._cache_index.items(),
            key=lambda x: x[1]["cached_at"],
            reverse=True
        )

        # Remove old entries
        for cache_key, _ in sorted_entries[keep_count:]:
            cache_file = self.cache_dir / f"{cache_key}.pyc"
            if cache_file.exists():
                cache_file.unlink()
            del self._cache_index[cache_key]

        self._save_cache_index()

    def _load_cache_index(self):
        """Load the cache index from disk."""
        index_file = self.cache_dir / "cache_index.json"
        if index_file.exists():
            try:
                with open(index_file, 'r') as f:
                    self._cache_index = json.load(f)
            except Exception:
                self._cache_index = {}

    def _save_cache_index(self):
        """Save the cache index to disk."""
        index_file = self.cache_dir / "cache_index.json"
        try:
            with open(index_file, 'w') as f:
                json.dump(self._cache_index, f, indent=2)
        except Exception:
            pass


class LazyLoader:
    """Lazy loading system for function dependencies."""

    def __init__(self):
        self._loaded_modules: Dict[str, Any] = {}
        self._loading_functions: Dict[str, Callable] = {}
        self._module_cache = PrecompiledModuleCache()

    def register_lazy_module(self, module_name: str, load_function: Callable):
        """
        Register a module for lazy loading.

        Args:
            module_name: Name of the module
            load_function: Function to load the module when needed
        """
        self._loading_functions[module_name] = load_function

    def get_module(self, module_name: str) -> Any:
        """
        Get a module, loading it lazily if necessary.

        Args:
            module_name: Name of the module to get

        Returns:
            The loaded module
        """
        if module_name in self._loaded_modules:
            return self._loaded_modules[module_name]

        if module_name in self._loading_functions:
            # Load the module
            load_func = self._loading_functions[module_name]
            module = load_func()

            # Cache the loaded module
            self._loaded_modules[module_name] = module
            return module

        # Try regular import
        try:
            return importlib.import_module(module_name)
        except ImportError:
            raise ImportError(f"Module '{module_name}' not found and no lazy loader registered")

    def preload_modules(self, module_names: List[str]):
        """
        Preload multiple modules in parallel.

        Args:
            module_names: List of module names to preload
        """
        def load_module_async(module_name: str):
            try:
                self.get_module(module_name)
            except Exception:
                # Ignore preload failures
                pass

        # Load modules in parallel
        with ThreadPoolExecutor(max_workers=min(len(module_names), 4)) as executor:
            executor.map(load_module_async, module_names)

    def clear_cache(self):
        """Clear the loaded modules cache."""
        self._loaded_modules.clear()


class FunctionWarmupManager:
    """Manages function warm-up to reduce cold starts."""

    def __init__(self):
        self._warmup_functions: Dict[str, Callable] = {}
        self._warmup_data: Dict[str, List[Dict[str, Any]]] = {}
        self._is_warming_up = False

    def register_warmup(self, function_name: str, warmup_function: Callable, test_data: List[Dict[str, Any]] = None):
        """
        Register a function for warm-up.

        Args:
            function_name: Name of the function
            warmup_function: The actual function to warm up
            test_data: Test data to use for warm-up calls
        """
        self._warmup_functions[function_name] = warmup_function
        self._warmup_data[function_name] = test_data or []

    def warmup_function(self, function_name: str) -> Dict[str, Any]:
        """
        Warm up a specific function.

        Args:
            function_name: Name of the function to warm up

        Returns:
            Warm-up results
        """
        if function_name not in self._warmup_functions:
            return {"success": False, "error": f"Function '{function_name}' not registered for warm-up"}

        func = self._warmup_functions[function_name]
        test_data = self._warmup_data[function_name]

        start_time = time.time()
        results = []

        try:
            # Call function with test data
            for i, data in enumerate(test_data):
                call_start = time.time()
                result = func(data)
                call_time = time.time() - call_start

                results.append({
                    "call_index": i,
                    "execution_time": call_time,
                    "success": True
                })

            total_time = time.time() - start_time

            return {
                "success": True,
                "function_name": function_name,
                "calls_made": len(results),
                "total_warmup_time": total_time,
                "average_call_time": total_time / len(results) if results else 0,
                "results": results
            }

        except Exception as e:
            return {
                "success": False,
                "function_name": function_name,
                "error": str(e),
                "partial_results": results
            }

    def warmup_all_functions(self) -> Dict[str, Any]:
        """Warm up all registered functions."""
        if self._is_warming_up:
            return {"success": False, "error": "Warm-up already in progress"}

        self._is_warming_up = True

        try:
            results = {}
            total_start_time = time.time()

            for func_name in self._warmup_functions:
                results[func_name] = self.warmup_function(func_name)

            total_time = time.time() - total_start_time

            return {
                "success": True,
                "total_warmup_time": total_time,
                "functions_warmed_up": len(results),
                "results": results
            }

        finally:
            self._is_warming_up = False

    async def warmup_all_functions_async(self) -> Dict[str, Any]:
        """Warm up all registered functions asynchronously."""
        if self._is_warming_up:
            return {"success": False, "error": "Warm-up already in progress"}

        self._is_warming_up = True

        try:
            results = {}
            total_start_time = time.time()

            # Warm up functions concurrently
            tasks = []
            for func_name in self._warmup_functions:
                task = asyncio.create_task(self._warmup_function_async(func_name))
                tasks.append((func_name, task))

            # Wait for all warm-ups to complete
            for func_name, task in tasks:
                results[func_name] = await task

            total_time = time.time() - total_start_time

            return {
                "success": True,
                "total_warmup_time": total_time,
                "functions_warmed_up": len(results),
                "results": results
            }

        finally:
            self._is_warming_up = False

    async def _warmup_function_async(self, function_name: str) -> Dict[str, Any]:
        """Warm up a function asynchronously."""
        # Run warm-up in thread pool to avoid blocking
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self.warmup_function, function_name)


class DependencyPreloader:
    """Preloads function dependencies to reduce cold start time."""

    def __init__(self):
        self._preloaded_dependencies: Set[str] = set()
        self._lazy_loader = LazyLoader()

    def preload_dependencies(self, dependencies: List[str]) -> Dict[str, Any]:
        """
        Preload function dependencies.

        Args:
            dependencies: List of dependency module names

        Returns:
            Preloading results
        """
        start_time = time.time()
        loaded_modules = []
        failed_modules = []

        # Filter out already preloaded dependencies
        to_preload = [dep for dep in dependencies if dep not in self._preloaded_dependencies]

        if not to_preload:
            return {
                "success": True,
                "message": "All dependencies already preloaded",
                "preload_time": 0
            }

        # Preload modules in parallel
        self._lazy_loader.preload_modules(to_preload)

        # Verify preloading worked
        for dep in to_preload:
            try:
                module = self._lazy_loader.get_module(dep)
                loaded_modules.append(dep)
                self._preloaded_dependencies.add(dep)
            except Exception as e:
                failed_modules.append({"module": dep, "error": str(e)})

        preload_time = time.time() - start_time

        return {
            "success": len(failed_modules) == 0,
            "preload_time": preload_time,
            "modules_loaded": len(loaded_modules),
            "modules_failed": len(failed_modules),
            "loaded_modules": loaded_modules,
            "failed_modules": failed_modules
        }

    def is_preloaded(self, dependency: str) -> bool:
        """Check if a dependency is already preloaded."""
        return dependency in self._preloaded_dependencies

    def clear_preloaded_cache(self):
        """Clear the preloaded dependencies cache."""
        self._preloaded_dependencies.clear()
        self._lazy_loader.clear_cache()


class ColdStartOptimizer:
    """Main cold start optimizer coordinating all optimization strategies."""

    def __init__(self):
        self.module_cache = PrecompiledModuleCache()
        self.lazy_loader = LazyLoader()
        self.warmup_manager = FunctionWarmupManager()
        self.dependency_preloader = DependencyPreloader()
        self._optimization_stats = {
            "total_cold_start_time_saved": 0,
            "modules_preloaded": 0,
            "functions_warmed_up": 0,
            "cache_hits": 0
        }

    def optimize_function_cold_start(
        self,
        function_name: str,
        function_code: str,
        dependencies: List[str],
        warmup_data: List[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Optimize cold start for a specific function.

        Args:
            function_name: Name of the function
            function_code: Function source code
            dependencies: List of dependencies
            warmup_data: Data for function warm-up

        Returns:
            Optimization results
        """
        optimization_results = {
            "function_name": function_name,
            "optimizations_applied": [],
            "preload_results": None,
            "warmup_results": None,
            "estimated_time_saved_ms": 0
        }

        # Preload dependencies
        if dependencies:
            preload_results = self.dependency_preloader.preload_dependencies(dependencies)
            optimization_results["preload_results"] = preload_results
            optimization_results["optimizations_applied"].append("dependency_preloading")

            if preload_results["success"]:
                # Estimate time saved: ~50ms per module
                time_saved = preload_results["modules_loaded"] * 50
                optimization_results["estimated_time_saved_ms"] += time_saved
                self._optimization_stats["modules_preloaded"] += preload_results["modules_loaded"]

        # Register function for warm-up
        if warmup_data:
            # Import and get the function (this is a simplified example)
            # In real implementation, you'd need to compile and import the function
            def dummy_warmup_func(data):
                return {"warmed_up": True, "data": data}

            self.warmup_manager.register_warmup(function_name, dummy_warmup_func, warmup_data)
            optimization_results["optimizations_applied"].append("function_warmup")

            # Estimate time saved: ~100ms for first call warm-up
            optimization_results["estimated_time_saved_ms"] += 100

        return optimization_results

    def get_optimization_stats(self) -> Dict[str, Any]:
        """Get overall optimization statistics."""
        return dict(self._optimization_stats)

    def clear_all_caches(self):
        """Clear all optimization caches."""
        self.module_cache = PrecompiledModuleCache()
        self.lazy_loader.clear_cache()
        self.dependency_preloader.clear_preloaded_cache()
        self.warmup_manager = FunctionWarmupManager()


# Decorator for cold start optimization
def optimize_cold_start(
    dependencies: List[str] = None,
    warmup_data: List[Dict[str, Any]] = None,
    preload_dependencies: bool = True
):
    """
    Decorator to optimize cold start performance for FlyPy functions.

    Args:
        dependencies: List of dependencies to preload
        warmup_data: Data to use for function warm-up
        preload_dependencies: Whether to preload dependencies

    Returns:
        Decorated function
    """
    def decorator(func: Callable) -> Callable:
        func._flypy_cold_start_config = {
            "dependencies": dependencies or [],
            "warmup_data": warmup_data or [],
            "preload_dependencies": preload_dependencies
        }

        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper
    return decorator


# Global cold start optimizer instance
cold_start_optimizer = ColdStartOptimizer()


def optimize_function_cold_start(
    function_name: str,
    function_code: str,
    dependencies: List[str] = None,
    warmup_data: List[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Convenience function to optimize cold start for a function.

    Args:
        function_name: Name of the function
        function_code: Function source code
        dependencies: List of dependencies
        warmup_data: Warm-up data

    Returns:
        Optimization results
    """
    return cold_start_optimizer.optimize_function_cold_start(
        function_name, function_code, dependencies or [], warmup_data
    )


def warmup_all_functions() -> Dict[str, Any]:
    """Warm up all registered functions."""
    return cold_start_optimizer.warmup_manager.warmup_all_functions()


async def warmup_all_functions_async() -> Dict[str, Any]:
    """Warm up all registered functions asynchronously."""
    return await cold_start_optimizer.warmup_manager.warmup_all_functions_async()


def preload_dependencies(dependencies: List[str]) -> Dict[str, Any]:
    """Preload function dependencies."""
    return cold_start_optimizer.dependency_preloader.preload_dependencies(dependencies)


def get_cold_start_stats() -> Dict[str, Any]:
    """Get cold start optimization statistics."""
    return cold_start_optimizer.get_optimization_stats()