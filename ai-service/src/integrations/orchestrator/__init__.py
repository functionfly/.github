"""Go Orchestrator Integration.

This module provides integration with the Go orchestrator API
for function management and execution.
"""

from .client import OrchestratorClient, get_orchestrator_client

__all__ = [
    "OrchestratorClient",
    "get_orchestrator_client",
]
