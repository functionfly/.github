"""
FlyPy - Deterministic Python Compilation

A Python SDK for compiling deterministic Python functions to WebAssembly
for execution on the FunctionFly platform.

Example:
    import flypy

    @flypy.function(
        name="calculate-total",
        deterministic=True,
        idempotent=True,
        cache_ttl=3600
    )
    def handler(event):
        '''Calculate order total with tax.'''
        items = event.get("items", [])
        tax_rate = event.get("tax_rate", 0.08)

        subtotal = sum(item["price"] * item["quantity"] for item in items)
        tax = subtotal * tax_rate
        total = subtotal + tax

        return {
            "subtotal": subtotal,
            "tax": tax,
            "total": total
        }
"""

from .decorators import (
    function,
    input_schema,
    output_schema,
    get_function_definition,
    get_registered_functions,
)
from .schema import Schema, Field
from .types import FunctionMetadata, ExecutionMode
from .optimizer import optimize_bundle, analyze_bundle_size
from .cold_start_optimizer import optimize_function_cold_start, warmup_all_functions
from .build_optimizer import optimize_build_process, clear_build_cache
from .performance_monitor import (
    monitor_performance,
    get_performance_stats,
    get_performance_report,
    start_performance_monitoring,
    stop_performance_monitoring,
    check_performance_alerts,
    start_performance_dashboard,
)
from .state import (
    StateClient,
    StateManager,
    StateError,
    StateNotFoundError,
    StatePermissionError,
    get_value,
    set_value,
    delete_value,
    get_history,
    create_snapshot,
    restore_snapshot,
    get_client,
)
from .edge_state import (
    EdgeStateClient,
    EdgeStateManager,
    EdgeStateError,
    EdgeStateNotFoundError,
    EdgeStatePermissionError,
    get as edge_get,
    set as edge_set,
    delete as edge_delete,
    snapshot as edge_snapshot,
)
from . import _functionfly
from ._functionfly import (
    get_attestation,
    delegate,
)
from .agent_types import (
    TrustPolicy,
    TrustedFunction,
    ToolExecutionMetadata,
    ToolExecutionEnvelope,
    AgentClientError,
    AgentHTTPError,
    TrustPolicyError,
)
from .agent_client import AgentClient
from .adapters import LangChainAdapter, AutoGenAdapter, CrewAIAdapter
from .version import __version__

__all__ = [
    "function",
    "input_schema",
    "output_schema",
    "get_function_definition",
    "get_registered_functions",
    "Schema",
    "Field",
    "FunctionMetadata",
    "ExecutionMode",
    "optimize_bundle",
    "analyze_bundle_size",
    "optimize_function_cold_start",
    "warmup_all_functions",
    "optimize_build_process",
    "clear_build_cache",
    "monitor_performance",
    "get_performance_stats",
    "get_performance_report",
    "start_performance_monitoring",
    "stop_performance_monitoring",
    "check_performance_alerts",
    "start_performance_dashboard",
    "StateClient",
    "StateManager",
    "StateError",
    "StateNotFoundError",
    "StatePermissionError",
    "get_value",
    "set_value",
    "delete_value",
    "get_history",
    "create_snapshot",
    "restore_snapshot",
    "get_client",
    "EdgeStateClient",
    "EdgeStateManager",
    "EdgeStateError",
    "EdgeStateNotFoundError",
    "EdgeStatePermissionError",
    "edge_get",
    "edge_set",
    "edge_delete",
    "edge_snapshot",
    "_functionfly",
    "get_attestation",
    "delegate",
    "TrustPolicy",
    "TrustedFunction",
    "ToolExecutionMetadata",
    "ToolExecutionEnvelope",
    "AgentClientError",
    "AgentHTTPError",
    "TrustPolicyError",
    "AgentClient",
    "LangChainAdapter",
    "AutoGenAdapter",
    "CrewAIAdapter",
    "__version__",
]
