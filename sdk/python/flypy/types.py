"""
Core types for FlyPy functions and metadata.
"""

from enum import Enum
from typing import Dict, List, Optional, Any, Union
from pydantic import BaseModel, Field as PydanticField


class ExecutionMode(str, Enum):
    """Execution modes for FlyPy functions."""

    DETERMINISTIC = "deterministic"
    COMPATIBLE = "compatible"


class FunctionMetadata(BaseModel):
    """Metadata for a FlyPy function."""

    name: str
    version: str = "1.0.0"
    description: Optional[str] = None

    # Execution properties
    deterministic: bool = True
    idempotent: bool = False
    pure: bool = False

    # Caching
    cache_ttl: Optional[int] = None

    # Capabilities and permissions
    capabilities: List[str] = PydanticField(default_factory=list)
    max_execution_time: Optional[int] = None  # in milliseconds

    # Schema information
    input_schema: Optional[Dict[str, Any]] = None
    output_schema: Optional[Dict[str, Any]] = None

    # Build information
    source_file: Optional[str] = None
    source_hash: Optional[str] = None

    # Runtime information
    execution_mode: ExecutionMode = ExecutionMode.DETERMINISTIC

    class Config:
        use_enum_values = True


class FunctionDefinition(BaseModel):
    """Complete function definition including code and metadata."""

    metadata: FunctionMetadata
    source_code: str
    ast_json: Optional[str] = None  # Serialized AST for the Go compiler

    # Derived information
    dependencies: List[str] = PydanticField(default_factory=list)
    imports: List[str] = PydanticField(default_factory=list)

    class Config:
        use_enum_values = True


class BuildResult(BaseModel):
    """Result of building a FlyPy function."""

    success: bool
    function_name: str
    output_dir: str
    warnings: List[str] = PydanticField(default_factory=list)
    errors: List[str] = PydanticField(default_factory=list)

    # Artifact information
    wasm_file: Optional[str] = None
    manifest_file: Optional[str] = None
    determinism_hash: Optional[str] = None

    # Performance metrics
    build_time_ms: Optional[int] = None
    wasm_size_bytes: Optional[int] = None

    # Optimization information
    optimization_stats: Dict[str, Any] = PydanticField(default_factory=dict)
    bundle_analysis: Dict[str, Any] = PydanticField(default_factory=dict)
    cold_start_stats: Dict[str, Any] = PydanticField(default_factory=dict)

    class Config:
        use_enum_values = True