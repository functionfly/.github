"""
EdgeStateClient - StateFabric access for edge functions.

A lightweight Python SDK module for accessing StateFabric from functions
running at the edge (WASM runtime). This module provides direct access to
state operations via WASM host functions, bypassing HTTP API calls.

Example (inside a FlyPy function running at edge):
    from flypy.edge_state import EdgeStateClient

    # Initialize client with fabric context
    state = EdgeStateClient(fabric_id="my-fabric")

    # Get state value
    cart = state.get("cart/user123")

    # Set state value
    state.set("cart/user123", {"items": [{"id": 1, "qty": 2}]})

    # Delete state value
    state.delete("cart/user123")

    # Create snapshot
    snapshot = state.snapshot("cart", label="pre-checkout")
"""

import json
import os
from typing import Any, Dict, List, Optional, Union


class EdgeStateError(Exception):
    """Base exception for edge state operations."""

    pass


class EdgeStateNotFoundError(EdgeStateError):
    """Raised when state is not found."""

    pass


class EdgeStatePermissionError(EdgeStateError):
    """Raised when permission is denied."""

    pass


# WASM host function imports - these are provided by the runtime
# When running outside WASM (e.g., local testing), these will be None


# Check if we're running in WASM environment
def _is_wasm_runtime() -> bool:
    """Check if running in WASM runtime environment."""
    # In WASM, certain modules/libraries won't be available
    try:
        import socket

        # If socket works fully (not just stub), we're likely not in WASM
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.close()
        return False
    except (OSError, ImportError, AttributeError):
        # Limited socket support suggests WASM environment
        return True


class EdgeStateClient:
    """
    Client for interacting with StateFabric from edge functions.

    This client uses WASM host functions for direct state access,
    providing low-latency state operations for functions running at
    the edge. When running outside WASM (local testing), it falls
    back to the standard HTTP-based StateClient.

    Attributes:
        fabric_id: The StateFabric identifier
        tenant_id: The tenant ID (auto-detected from env if not provided)
        use_wasm: Whether to use WASM host functions (auto-detected)
    """

    def __init__(
        self,
        fabric_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
        use_wasm: Optional[bool] = None,
    ):
        """
        Initialize the edge state client.

        Args:
            fabric_id: The StateFabric identifier (required for most operations)
            tenant_id: The tenant ID (defaults to FLYPY_TENANT_ID env var)
            use_wasm: Force WASM mode (auto-detected if not specified)
        """
        self.fabric_id = fabric_id
        self.tenant_id = tenant_id or os.environ.get("FLYPY_TENANT_ID", "default")
        self._use_wasm = use_wasm if use_wasm is not None else _is_wasm_runtime()
        self._http_client = None

        # Import StateClient for fallback mode
        if not self._use_wasm:
            from .state import StateClient

            self._http_client = StateClient(
                tenant_id=self.tenant_id, api_key=os.environ.get("FLYPY_API_KEY")
            )

    def _get_full_path(self, key: str) -> str:
        """Build the full state path including tenant and fabric."""
        if self.fabric_id:
            return f"{self.tenant_id}/{self.fabric_id}/{key}"
        return f"{self.tenant_id}/{key}"

    def _call_wasm_host(self, func_name: str, *args) -> Any:
        """
        Call a WASM host function.

        In a real WASM environment, these would be imported from
        the host. For now, we use a placeholder implementation.
        """
        # TODO: When WASM Python runtime supports host function imports,
        # this will be replaced with actual WASM import calls
        # Example: return __import__("functionfly").state_get(*args)
        raise EdgeStateError(
            f"WASM host function '{func_name}' not available. "
            "Ensure you're running in the FunctionFly WASM runtime."
        )

    def get(self, key: str, default: Any = None) -> Any:
        """
        Get a value from state.

        Args:
            key: The state key (e.g., "cart/user123")
            default: Default value if key not found

        Returns:
            The state value, or default if not found
        """
        path = self._get_full_path(key)

        if self._use_wasm:
            try:
                # Call WASM host function: functionfly.state_get
                result = self._call_wasm_host("state_get", path)
                data = json.loads(result)
                return data.get("value", default)
            except (EdgeStateError, json.JSONDecodeError):
                return default
        else:
            # Fallback to HTTP client
            try:
                result = self._http_client.get_value(path)
                if result and "value" in result:
                    return result["value"]
            except Exception:
                pass
            return default

    def set(
        self, key: str, value: Any, metadata: Optional[Dict[str, Any]] = None
    ) -> bool:
        """
        Set a value in state.

        Args:
            key: The state key
            value: The value to store (must be JSON serializable)
            metadata: Optional metadata to store with the value

        Returns:
            True if successful, False otherwise
        """
        path = self._get_full_path(key)

        try:
            data = {"value": value}
            if metadata:
                data["metadata"] = metadata
            json_data = json.dumps(data)
        except (TypeError, ValueError) as e:
            raise EdgeStateError(f"Value must be JSON serializable: {e}")

        if self._use_wasm:
            try:
                # Call WASM host function: functionfly.state_set
                self._call_wasm_host("state_set", path, json_data)
                return True
            except EdgeStateError:
                return False
        else:
            # Fallback to HTTP client
            try:
                self._http_client.set_value(path, value, metadata)
                return True
            except Exception:
                return False

    def delete(self, key: str) -> bool:
        """
        Delete a value from state.

        Args:
            key: The state key

        Returns:
            True if successful, False otherwise
        """
        path = self._get_full_path(key)

        if self._use_wasm:
            try:
                # Call WASM host function: functionfly.state_delete
                self._call_wasm_host("state_delete", path)
                return True
            except EdgeStateError:
                return False
        else:
            # Fallback to HTTP client
            try:
                self._http_client.delete_value(path)
                return True
            except Exception:
                return False

    def get_all(self, prefix: str = "") -> Dict[str, Any]:
        """
        Get all values with a given prefix.

        Args:
            prefix: Key prefix to filter by

        Returns:
            Dictionary of key-value pairs
        """
        # Note: This operation may be expensive and should be used carefully
        path = self._get_full_path(prefix) if prefix else self._get_full_path("")

        if self._use_wasm:
            # WASM mode: get_fabric returns all state in fabric
            try:
                result = self._call_wasm_host("state_get_fabric", self.fabric_id or "")
                data = json.loads(result)
                return data.get("values", {})
            except (EdgeStateError, json.JSONDecodeError):
                return {}
        else:
            # Fallback to HTTP client
            try:
                return self._http_client.get_all_values(path)
            except Exception:
                return {}

    def snapshot(
        self, key: str, label: Optional[str] = None
    ) -> Optional[Dict[str, Any]]:
        """
        Create a snapshot of a state key.

        Args:
            key: The state key to snapshot
            label: Optional label for the snapshot

        Returns:
            Snapshot metadata, or None if failed
        """
        path = self._get_full_path(key)

        if self._use_wasm:
            try:
                result = self._call_wasm_host(
                    "state_create_snapshot", path, label or ""
                )
                return json.loads(result)
            except (EdgeStateError, json.JSONDecodeError):
                return None
        else:
            # Fallback to HTTP client
            try:
                return self._http_client.create_snapshot(path, label)
            except Exception:
                return None

    def get_fabric_info(self) -> Optional[Dict[str, Any]]:
        """
        Get information about the current fabric.

        Returns:
            Fabric metadata, or None if failed
        """
        if not self.fabric_id:
            return None

        if self._use_wasm:
            try:
                result = self._call_wasm_host("state_get_fabric", self.fabric_id)
                return json.loads(result)
            except (EdgeStateError, json.JSONDecodeError):
                return None
        else:
            # HTTP client doesn't have direct fabric info method
            return {"fabric_id": self.fabric_id, "tenant_id": self.tenant_id}


class EdgeStateManager:
    """
    State manager for declarative state access in edge functions.

    Provides a decorator-based interface for managing state within
    edge functions, similar to StateManager but optimized for
    WASM runtime.

    Example:
        from flypy.edge_state import EdgeStateManager

        manager = EdgeStateManager(fabric_id="my-fabric")

        @manager.state("cart", write=True)
        def get_cart(user_id: str) -> dict:
            pass

        @manager.state("cart", write=True)
        def update_cart(user_id: str, item: dict) -> dict:
            pass
    """

    def __init__(
        self,
        fabric_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
        use_wasm: Optional[bool] = None,
    ):
        """
        Initialize the edge state manager.

        Args:
            fabric_id: The StateFabric identifier
            tenant_id: The tenant ID
            use_wasm: Force WASM mode (auto-detected if not specified)
        """
        self.client = EdgeStateClient(fabric_id, tenant_id, use_wasm)

    def state(
        self,
        state_name: str,
        key: Optional[str] = None,
        write: bool = False,
        create_if_not_exists: bool = True,
    ):
        """
        Decorator for accessing state in an edge function.

        Args:
            state_name: Name of the state container/key prefix
            key: Optional key (can be function param)
            write: Whether this is a write operation
            create_if_not_exists: Create state if it doesn't exist

        Returns:
            Decorator function
        """
        from functools import wraps

        def decorator(func):
            @wraps(func)
            def wrapper(*args, **kwargs):
                # Determine the key
                actual_key = key
                if not actual_key and args:
                    actual_key = str(args[0])

                full_key = f"{state_name}/{actual_key}" if actual_key else state_name

                if write:
                    # Call the function and store result
                    result = func(*args, **kwargs)
                    self.client.set(full_key, result)
                    return result
                else:
                    # Get the value and pass to function
                    value = self.client.get(full_key)
                    if value is not None:
                        return value
                    # If no value, call function and store result
                    result = func(*args, **kwargs)
                    if create_if_not_exists:
                        self.client.set(full_key, result)
                    return result

            return wrapper

        return decorator

    def get(self, state_name: str, key: str, default: Any = None) -> Any:
        """Get a value from state."""
        full_key = f"{state_name}/{key}"
        return self.client.get(full_key, default)

    def set(self, state_name: str, key: str, value: Any) -> bool:
        """Set a value in state."""
        full_key = f"{state_name}/{key}"
        return self.client.set(full_key, value)

    def delete(self, state_name: str, key: str) -> bool:
        """Delete a value from state."""
        full_key = f"{state_name}/{key}"
        return self.client.delete(full_key)


# Convenience functions using default client


def _get_default_client() -> EdgeStateClient:
    """Get or create the default edge state client."""
    fabric_id = os.environ.get("FLYPY_STATE_FABRIC_ID")
    tenant_id = os.environ.get("FLYPY_TENANT_ID")
    return EdgeStateClient(fabric_id=fabric_id, tenant_id=tenant_id)


def get(key: str, default: Any = None, fabric_id: Optional[str] = None) -> Any:
    """Get a value from state using the default client."""
    client = _get_default_client()
    if fabric_id:
        client.fabric_id = fabric_id
    return client.get(key, default)


def set(key: str, value: Any, fabric_id: Optional[str] = None) -> bool:
    """Set a value in state using the default client."""
    client = _get_default_client()
    if fabric_id:
        client.fabric_id = fabric_id
    return client.set(key, value)


def delete(key: str, fabric_id: Optional[str] = None) -> bool:
    """Delete a value from state using the default client."""
    client = _get_default_client()
    if fabric_id:
        client.fabric_id = fabric_id
    return client.delete(key)


def snapshot(
    key: str, label: Optional[str] = None, fabric_id: Optional[str] = None
) -> Optional[Dict[str, Any]]:
    """Create a snapshot using the default client."""
    client = _get_default_client()
    if fabric_id:
        client.fabric_id = fabric_id
    return client.snapshot(key, label)


__all__ = [
    "EdgeStateClient",
    "EdgeStateManager",
    "EdgeStateError",
    "EdgeStateNotFoundError",
    "EdgeStatePermissionError",
    "get",
    "set",
    "delete",
    "snapshot",
]
