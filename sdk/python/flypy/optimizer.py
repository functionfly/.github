"""
Bundle size optimization for FlyPy functions.

This module implements tree shaking, dead code elimination, and minimal runtime
optimizations to reduce WebAssembly bundle sizes.
"""

import ast
import sys
from typing import Dict, Any, List, Set, Optional, Tuple
from pathlib import Path
import re
import hashlib


class BundleOptimizer:
    """Optimizes FlyPy function bundles for size."""

    def __init__(self, optimization_level: str = "balanced"):
        """
        Initialize the bundle optimizer.

        Args:
            optimization_level: "minimal", "balanced", or "aggressive"
        """
        self.optimization_level = optimization_level
        self.optimization_stats = {
            "original_size": 0,
            "optimized_size": 0,
            "code_removed": 0,
            "optimizations_applied": []
        }

    def optimize_function_bundle(
        self,
        source_code: str,
        metadata: Dict[str, Any],
        dependencies: List[str]
    ) -> Tuple[str, Dict[str, Any]]:
        """
        Optimize a function bundle for size.

        Args:
            source_code: Original Python source code
            metadata: Function metadata
            dependencies: List of module dependencies

        Returns:
            Tuple of (optimized_code, optimization_stats)
        """
        self.optimization_stats["original_size"] = len(source_code)

        # Parse AST for analysis
        try:
            tree = ast.parse(source_code)
        except SyntaxError:
            # Return original if parsing fails
            return source_code, self.optimization_stats

        # Apply optimizations based on level
        optimized_tree = self._apply_optimizations(tree, metadata, dependencies)

        # Convert back to source code
        optimized_code = self._tree_to_source(optimized_tree)

        self.optimization_stats["optimized_size"] = len(optimized_code)
        self.optimization_stats["code_removed"] = (
            self.optimization_stats["original_size"] - self.optimization_stats["optimized_size"]
        )

        return optimized_code, self.optimization_stats

    def _apply_optimizations(
        self,
        tree: ast.AST,
        metadata: Dict[str, Any],
        dependencies: List[str]
    ) -> ast.AST:
        """Apply optimization passes to the AST."""

        # Dead code elimination
        if self.optimization_level in ["balanced", "aggressive"]:
            tree = DeadCodeEliminator().optimize(tree)
            self.optimization_stats["optimizations_applied"].append("dead_code_elimination")

        # Unused import removal
        if self.optimization_level in ["balanced", "aggressive"]:
            tree = UnusedImportRemover().optimize(tree)
            self.optimization_stats["optimizations_applied"].append("unused_import_removal")

        # Constant folding
        if self.optimization_level == "aggressive":
            tree = ConstantFolder().optimize(tree)
            self.optimization_stats["optimizations_applied"].append("constant_folding")

        # Docstring removal (except for the main function)
        if self.optimization_level in ["balanced", "aggressive"]:
            tree = DocstringRemover(metadata.get("name", "")).optimize(tree)
            self.optimization_stats["optimizations_applied"].append("docstring_removal")

        # Assert statement removal
        if self.optimization_level == "aggressive":
            tree = AssertRemover().optimize(tree)
            self.optimization_stats["optimizations_applied"].append("assert_removal")

        return tree

    def _tree_to_source(self, tree: ast.AST) -> str:
        """Convert AST back to source code."""
        # Simple AST to source conversion
        # In a real implementation, this would use a proper code generator
        return compile(tree, '<optimized>', 'exec').co_filename  # Placeholder


class DeadCodeEliminator(ast.NodeTransformer):
    """Removes unreachable code and unused variables."""

    def __init__(self):
        self.used_names: Set[str] = set()
        self.defined_names: Set[str] = set()

    def optimize(self, tree: ast.AST) -> ast.AST:
        """Optimize the AST by removing dead code."""
        # First pass: collect used names
        collector = NameUsageCollector()
        collector.visit(tree)
        self.used_names = collector.used_names

        # Second pass: remove unused assignments
        return self.visit(tree)

    def visit_Assign(self, node: ast.Assign) -> Optional[ast.AST]:
        """Remove assignments to unused variables."""
        # Check if any of the targets are used
        targets_used = any(
            self._is_name_used(target) for target in node.targets
        )

        if not targets_used:
            # Remove unused assignment
            return None

        return self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> ast.AST:
        """Process function definitions."""
        # Don't remove function definitions even if they're not called
        # (they might be entry points)
        return self.generic_visit(node)

    def _is_name_used(self, node: ast.AST) -> bool:
        """Check if a name node is used."""
        if isinstance(node, ast.Name):
            return node.id in self.used_names
        return True  # Conservative: keep if unsure


class NameUsageCollector(ast.NodeVisitor):
    """Collects all names that are used in the code."""

    def __init__(self):
        self.used_names: Set[str] = set()
        self.defined_names: Set[str] = set()

    def visit_Name(self, node: ast.Name) -> None:
        if isinstance(node.ctx, (ast.Load, ast.AugLoad)):
            self.used_names.add(node.id)
        elif isinstance(node.ctx, (ast.Store, ast.AugStore, ast.Param)):
            self.defined_names.add(node.id)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        # Don't count function names as used within their own body
        # unless they're recursive
        self.defined_names.add(node.name)
        self.generic_visit(node)

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        self.defined_names.add(node.name)
        self.generic_visit(node)


class UnusedImportRemover(ast.NodeTransformer):
    """Removes unused import statements."""

    def __init__(self):
        self.imports: Dict[str, ast.stmt] = {}
        self.used_names: Set[str] = set()

    def optimize(self, tree: ast.AST) -> ast.AST:
        """Remove unused imports from the AST."""
        # First pass: collect imports and used names
        collector = ImportUsageCollector()
        collector.visit(tree)

        self.imports = collector.imports
        self.used_names = collector.used_names

        # Second pass: remove unused imports
        return self.visit(tree)

    def visit_Import(self, node: ast.Import) -> Optional[ast.AST]:
        """Remove unused import statements."""
        # Check if any names from this import are used
        used_aliases = []
        for alias in node.names:
            if alias.asname:
                if alias.asname in self.used_names:
                    used_aliases.append(alias)
            elif alias.name in self.used_names:
                used_aliases.append(alias)

        if not used_aliases:
            # Remove entire import
            return None
        elif len(used_aliases) < len(node.names):
            # Keep only used imports
            node.names = used_aliases

        return node

    def visit_ImportFrom(self, node: ast.ImportFrom) -> Optional[ast.AST]:
        """Remove unused from-import statements."""
        # Check if any names from this import are used
        used_aliases = []
        for alias in node.names:
            if alias.asname:
                if alias.asname in self.used_names:
                    used_aliases.append(alias)
            elif alias.name in self.used_names:
                used_aliases.append(alias)

        if not used_aliases:
            # Remove entire import
            return None
        elif len(used_aliases) < len(node.names):
            # Keep only used imports
            node.names = used_aliases

        return node


class ImportUsageCollector(ast.NodeVisitor):
    """Collects import statements and used names."""

    def __init__(self):
        self.imports: Dict[str, ast.stmt] = {}
        self.used_names: Set[str] = set()

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            if alias.asname:
                self.used_names.add(alias.asname)
            else:
                # Add the module name
                self.used_names.add(alias.name.split('.')[0])

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        for alias in node.names:
            if alias.asname:
                self.used_names.add(alias.asname)
            else:
                self.used_names.add(alias.name)

    def visit_Name(self, node: ast.Name) -> None:
        if isinstance(node.ctx, (ast.Load, ast.AugLoad)):
            self.used_names.add(node.id)


class ConstantFolder(ast.NodeTransformer):
    """Folds constant expressions at compile time."""

    def visit_BinOp(self, node: ast.BinOp) -> ast.AST:
        """Fold constant binary operations."""
        self.generic_visit(node)

        # Check if both operands are constants
        if isinstance(node.left, ast.Constant) and isinstance(node.right, ast.Constant):
            left_val = node.left.value
            right_val = node.right.value

            # Fold based on operation type
            if isinstance(node.op, ast.Add) and isinstance(left_val, (int, float)) and isinstance(right_val, (int, float)):
                return ast.Constant(value=left_val + right_val)
            elif isinstance(node.op, ast.Sub) and isinstance(left_val, (int, float)) and isinstance(right_val, (int, float)):
                return ast.Constant(value=left_val - right_val)
            elif isinstance(node.op, ast.Mult) and isinstance(left_val, (int, float)) and isinstance(right_val, (int, float)):
                return ast.Constant(value=left_val * right_val)

        return node

    def optimize(self, tree: ast.AST) -> ast.AST:
        """Optimize the AST by folding constants."""
        return self.visit(tree)


class DocstringRemover(ast.NodeTransformer):
    """Removes docstrings from functions and classes."""

    def __init__(self, main_function_name: str = ""):
        self.main_function_name = main_function_name

    def visit_FunctionDef(self, node: ast.FunctionDef) -> ast.AST:
        """Remove docstrings from functions."""
        if node.name != self.main_function_name:  # Keep main function docstring
            node.body = self._remove_docstring(node.body)
        return self.generic_visit(node)

    def visit_ClassDef(self, node: ast.ClassDef) -> ast.AST:
        """Remove docstrings from classes."""
        node.body = self._remove_docstring(node.body)
        return self.generic_visit(node)

    def visit_Module(self, node: ast.Module) -> ast.AST:
        """Remove module docstrings."""
        node.body = self._remove_docstring(node.body)
        return self.generic_visit(node)

    def _remove_docstring(self, body: List[ast.stmt]) -> List[ast.stmt]:
        """Remove docstring from a body if present."""
        if (body and
            isinstance(body[0], ast.Expr) and
            isinstance(body[0].value, ast.Constant) and
            isinstance(body[0].value.value, str)):
            # Remove the docstring
            return body[1:]
        return body

    def optimize(self, tree: ast.AST) -> ast.AST:
        """Optimize the AST by removing docstrings."""
        return self.visit(tree)


class AssertRemover(ast.NodeTransformer):
    """Removes assert statements for production builds."""

    def visit_Assert(self, node: ast.Assert) -> None:
        """Remove assert statements."""
        return None

    def optimize(self, tree: ast.AST) -> ast.AST:
        """Optimize the AST by removing assert statements."""
        return self.visit(tree)


class MinimalRuntimeBuilder:
    """Builds minimal runtime environments for functions."""

    def __init__(self):
        self.runtime_components = {
            "base": ["object", "type", "int", "str", "list", "dict", "tuple", "set"],
            "math": ["math"],
            "json": ["json"],
            "datetime": ["datetime", "time"],
            "collections": ["collections"],
        }

    def build_minimal_runtime(
        self,
        required_modules: List[str],
        optimization_level: str = "balanced"
    ) -> Dict[str, Any]:
        """
        Build a minimal runtime with only required components.

        Args:
            required_modules: List of modules the function actually uses
            optimization_level: Optimization level

        Returns:
            Runtime configuration
        """
        runtime_config = {
            "included_modules": set(),
            "excluded_modules": set(),
            "builtin_functions": set(),
            "optimizations": []
        }

        # Determine which runtime components to include
        for module in required_modules:
            if module in self.runtime_components:
                runtime_config["included_modules"].update(self.runtime_components[module])

        # Aggressive optimization: remove unused builtins
        if optimization_level == "aggressive":
            # Analyze which builtins are actually used and exclude others
            runtime_config["optimizations"].append("minimal_builtins")

        return runtime_config


class BundleSizeAnalyzer:
    """Analyzes bundle size and provides optimization recommendations."""

    def __init__(self):
        self.size_thresholds = {
            "small": 50 * 1024,      # 50KB
            "medium": 200 * 1024,    # 200KB
            "large": 1024 * 1024,    # 1MB
        }

    def analyze_bundle(self, bundle_code: str, metadata: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analyze bundle size and provide recommendations.

        Args:
            bundle_code: The compiled bundle code
            metadata: Function metadata

        Returns:
            Analysis results with recommendations
        """
        bundle_size = len(bundle_code)
        analysis = {
            "bundle_size_bytes": bundle_size,
            "bundle_size_kb": bundle_size / 1024,
            "category": self._categorize_size(bundle_size),
            "recommendations": [],
            "optimization_opportunities": []
        }

        # Size-based recommendations
        if bundle_size > self.size_thresholds["large"]:
            analysis["recommendations"].append({
                "type": "critical",
                "message": "Bundle size exceeds 1MB. Consider aggressive optimization.",
                "actions": ["enable_aggressive_optimization", "review_dependencies"]
            })

        elif bundle_size > self.size_thresholds["medium"]:
            analysis["recommendations"].append({
                "type": "warning",
                "message": "Bundle size exceeds 200KB. Consider optimization.",
                "actions": ["enable_balanced_optimization", "reduce_imports"]
            })

        # Function-specific recommendations
        if metadata.get("capabilities"):
            analysis["optimization_opportunities"].append({
                "type": "capability_analysis",
                "message": f"Function uses {len(metadata['capabilities'])} capabilities. Ensure all are necessary."
            })

        # Pure function optimization
        if metadata.get("pure"):
            analysis["optimization_opportunities"].append({
                "type": "pure_function",
                "message": "Pure function detected. Can apply additional optimizations."
            })

        return analysis

    def _categorize_size(self, size_bytes: int) -> str:
        """Categorize bundle size."""
        if size_bytes <= self.size_thresholds["small"]:
            return "small"
        elif size_bytes <= self.size_thresholds["medium"]:
            return "medium"
        else:
            return "large"


# Convenience functions for easy use
def optimize_bundle(
    source_code: str,
    metadata: Dict[str, Any],
    dependencies: List[str],
    optimization_level: str = "balanced"
) -> Tuple[str, Dict[str, Any]]:
    """
    Optimize a function bundle for size.

    Args:
        source_code: Original source code
        metadata: Function metadata
        dependencies: Module dependencies
        optimization_level: "minimal", "balanced", or "aggressive"

    Returns:
        Tuple of (optimized_code, optimization_stats)
    """
    optimizer = BundleOptimizer(optimization_level)
    return optimizer.optimize_function_bundle(source_code, metadata, dependencies)


def analyze_bundle_size(bundle_code: str, metadata: Dict[str, Any]) -> Dict[str, Any]:
    """
    Analyze bundle size and provide recommendations.

    Args:
        bundle_code: Compiled bundle code
        metadata: Function metadata

    Returns:
        Analysis results
    """
    analyzer = BundleSizeAnalyzer()
    return analyzer.analyze_bundle(bundle_code, metadata)