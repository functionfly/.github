"""
StateFabric - Composable Durable State for Stateless Functions

A Python SDK module for accessing and managing durable state bound to function identities.

Example:
    import flypy
    from flypy.state import StateClient, StateManager

    # Direct client usage
    client = StateClient()

    # Get state value
    cart = client.get_value("my-tenant/cart/user123")
    print(cart)

    # Set state value
    client.set_value("my-tenant/cart/user123", {"items": [{"id": 1, "qty": 2}]})

    # Get state history
    history = client.get_history("my-tenant/cart/user123")

    # Create snapshot
    snapshot = client.create_snapshot("my-tenant/cart/user123", label="backup-001")

    # Restore from snapshot
    client.restore_snapshot("my-tenant/cart/user123", snapshot_version=1)

    # Using the state manager decorator
    manager = StateManager()

    @manager.state("cart", write=True)
    def get_cart(user_id: str) -> dict:
        '''Get user's cart state.'''
        pass

    @manager.state("cart", write=True)
    def update_cart(user_id: str, item: dict) -> dict:
        '''Update user's cart state.'''
        pass

    # Usage
    cart = get_cart("user123")
    updated_cart = update_cart("user123", {"id": 1, "qty": 2})
"""

import json
import os
import urllib.request
import urllib.error
from typing import Any, Dict, List, Optional
from datetime import datetime
from functools import wraps


# Configuration
DEFAULT_API_URL = os.environ.get("FLYPY_API_URL", "")  # Must be set explicitly
DEFAULT_TENANT_ID = os.environ.get("FLYPY_TENANT_ID", "")


class StateError(Exception):
    """Base exception for state operations."""
    pass


class StateNotFoundError(StateError):
    """Raised when state is not found."""
    pass


class StatePermissionError(StateError):
    """Raised when permission is denied."""
    pass


class StateClient:
    """
    Client for interacting with StateFabric API.

    Provides methods for CRUD operations on state, value management,
    history tracking, snapshots, and permissions.
    """

    def __init__(
        self,
        api_url: str = DEFAULT_API_URL,
        tenant_id: str = DEFAULT_TENANT_ID,
        api_key: Optional[str] = None
    ):
        """
        Initialize the state client.

        Args:
            api_url: Base URL for the FunctionFly API
            tenant_id: Tenant ID for multi-tenancy
            api_key: API key for authentication
        """
        self.api_url = api_url.rstrip("/")
        self.tenant_id = tenant_id
        self.api_key = api_key or os.environ.get("FLYPY_API_KEY", "")

    def _get_headers(self) -> Dict[str, str]:
        """Get request headers."""
        headers = {
            "Content-Type": "application/json",
            "Accept": "application/json"
        }
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers

    def _make_request(
        self,
        method: str,
        path: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Make an HTTP request to the API."""
        url = f"{self.api_url}{path}"

        if params:
            query_string = "&".join(f"{k}={v}" for k, v in params.items() if v is not None)
            if query_string:
                url += f"?{query_string}"

        body = json.dumps(data).encode("utf-8") if data else None

        request = urllib.request.Request(
            url,
            data=body,
            headers=self._get_headers(),
            method=method
        )

        try:
            with urllib.request.urlopen(request) as response:
                response_body = response.read().decode("utf-8")
                if response_body:
                    return json.loads(response_body)
                return {}
        except urllib.error.HTTPError as e:
            error_body = e.read().decode("utf-8")
            try:
                error_data = json.loads(error_body)
                error_message = error_data.get("error", str(e))
            except json.JSONDecodeError:
                error_message = error_body or str(e)

            if e.code == 404:
                raise StateNotFoundError(error_message)
            elif e.code == 403:
                raise StatePermissionError(error_message)
            else:
                raise StateError(f"HTTP {e.code}: {error_message}")
        except urllib.error.URLError as e:
            raise StateError(f"Network error: {e.reason}")

    # State Management

    def create_state(
        self,
        path: str,
        name: Optional[str] = None,
        storage_type: str = "durable",
        ttl_days: Optional[int] = None,
        max_size_mb: Optional[int] = None,
        is_versioned: bool = True,
        is_public: bool = False,
        description: Optional[str] = None,
        tags: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Create a new state container.

        Args:
            path: Full path including tenant and state name (e.g., "tenant/state-name")
            name: Optional name (defaults to path)
            storage_type: Type of storage (durable, ephemeral, cached)
            ttl_days: Time to live in days
            max_size_mb: Maximum size in MB
            is_versioned: Whether to version the state
            is_public: Whether state is publicly accessible
            description: Optional description
            tags: Optional tags

        Returns:
            Created state object
        """
        name = name or path.split("/")[-1]

        data = {
            "name": name,
            "storage_type": storage_type,
            "is_versioned": is_versioned,
            "is_public": is_public,
        }

        if ttl_days is not None:
            data["ttl_days"] = ttl_days
        if max_size_mb is not None:
            data["max_size_mb"] = max_size_mb
        if description:
            data["description"] = description
        if tags:
            data["tags"] = tags

        return self._make_request("POST", "/v1/state", data=data)

    def get_state(self, path: str) -> Dict[str, Any]:
        """
        Get a state container by path.

        Args:
            path: Full path including tenant and state name

        Returns:
            State object
        """
        return self._make_request("GET", f"/v1/state/{path}")

    def list_states(
        self,
        tenant_id: Optional[str] = None,
        limit: int = 100,
        offset: int = 0
    ) -> List[Dict[str, Any]]:
        """
        List all states for a tenant.

        Args:
            tenant_id: Tenant ID (defaults to client tenant_id)
            limit: Maximum number of results
            offset: Offset for pagination

        Returns:
            List of state objects
        """
        tenant_id = tenant_id or self.tenant_id
        params = {
            "tenant_id": tenant_id,
            "limit": limit,
            "offset": offset
        }
        return self._make_request("GET", "/v1/state", params=params)

    def delete_state(self, path: str) -> None:
        """
        Delete a state container.

        Args:
            path: Full path including tenant and state name
        """
        self._make_request("DELETE", f"/v1/state/{path}")

    # Value Operations

    def set_value(
        self,
        path: str,
        value: Dict[str, Any],
        metadata: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Set a value in the state.

        Args:
            path: Full path including tenant/state/key
            value: Value to set (JSON serializable dict)
            metadata: Optional metadata

        Returns:
            Updated value object
        """
        data = {"value": value}
        if metadata:
            data["metadata"] = metadata

        return self._make_request("PUT", f"/v1/state/{path}/value", data=data)

    def get_value(
        self,
        path: str,
        include_metadata: bool = False
    ) -> Optional[Dict[str, Any]]:
        """
        Get a value from the state.

        Args:
            path: Full path including tenant/state/key
            include_metadata: Whether to include metadata

        Returns:
            Value object or None if not found
        """
        params = {"include_metadata": include_metadata}
        return self._make_request("GET", f"/v1/state/{path}/value", params=params)

    def delete_value(self, path: str) -> None:
        """
        Delete a value from the state.

        Args:
            path: Full path including tenant/state/key
        """
        self._make_request("DELETE", f"/v1/state/{path}/value")

    def get_all_values(self, state_path: str) -> Dict[str, Any]:
        """
        Get all key-value pairs in a state container.

        Args:
            state_path: Full path including tenant/state

        Returns:
            Dictionary of all key-value pairs
        """
        return self._make_request("GET", f"/v1/state/{state_path}")

    # History & Events

    def get_history(
        self,
        path: str,
        limit: int = 100,
        offset: int = 0,
        event_types: Optional[List[str]] = None
    ) -> List[Dict[str, Any]]:
        """
        Get the event history for a state path.

        Args:
            path: Full path including tenant/state/key
            limit: Maximum number of events
            offset: Offset for pagination
            event_types: Filter by event types

        Returns:
            List of event objects
        """
        params = {
            "limit": limit,
            "offset": offset,
            "event_types": ",".join(event_types) if event_types else None
        }
        return self._make_request("GET", f"/v1/state/{path}/history", params=params)

    def time_travel(
        self,
        path: str,
        timestamp: Optional[datetime] = None,
        version: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Query state at a specific point in time or version.

        Args:
            path: Full path including tenant/state/key
            timestamp: ISO 8601 timestamp
            version: Specific version number

        Returns:
            State value at the specified time/version
        """
        params = {}
        if timestamp:
            params["timestamp"] = timestamp.isoformat()
        if version is not None:
            params["version"] = version

        return self._make_request("GET", f"/v1/state/{path}/time-travel", params=params)

    # Snapshots

    def create_snapshot(
        self,
        path: str,
        label: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Create a snapshot of the current state.

        Args:
            path: Full path including tenant/state
            label: Optional label for the snapshot

        Returns:
            Created snapshot object
        """
        data = {}
        if label:
            data["label"] = label

        return self._make_request("POST", f"/v1/state/{path}/snapshot", data=data)

    def list_snapshots(
        self,
        path: str,
        limit: int = 50,
        offset: int = 0
    ) -> List[Dict[str, Any]]:
        """
        List all snapshots for a state container.

        Args:
            path: Full path including tenant/state
            limit: Maximum number of snapshots
            offset: Offset for pagination

        Returns:
            List of snapshot objects
        """
        params = {"limit": limit, "offset": offset}
        return self._make_request("GET", f"/v1/state/{path}/snapshots", params=params)

    def restore_snapshot(
        self,
        path: str,
        snapshot_version: int
    ) -> Dict[str, Any]:
        """
        Restore state from a snapshot.

        Args:
            path: Full path including tenant/state
            snapshot_version: Version number of snapshot to restore

        Returns:
            Restored state object
        """
        data = {"snapshot_version": snapshot_version}
        return self._make_request("POST", f"/v1/state/{path}/restore", data=data)

    # Permissions

    def grant_permission(
        self,
        path: str,
        principal_type: str,
        principal_id: str,
        can_read: bool = True,
        can_write: bool = False,
        can_delete: bool = False,
        can_admin: bool = False,
        can_trigger: bool = False
    ) -> Dict[str, Any]:
        """
        Grant permission to access a state.

        Args:
            path: Full path including tenant/state
            principal_type: Type of principal (user, team, service)
            principal_id: ID of the principal
            can_read: Grant read permission
            can_write: Grant write permission
            can_delete: Grant delete permission
            can_admin: Grant admin permission
            can_trigger: Grant trigger permission

        Returns:
            Created permission object
        """
        data = {
            "principal_type": principal_type,
            "principal_id": principal_id,
            "can_read": can_read,
            "can_write": can_write,
            "can_delete": can_delete,
            "can_admin": can_admin,
            "can_trigger": can_trigger
        }

        return self._make_request("POST", f"/v1/state/{path}/permissions", data=data)

    def get_permissions(self, path: str) -> List[Dict[str, Any]]:
        """
        Get all permissions for a state.

        Args:
            path: Full path including tenant/state

        Returns:
            List of permission objects
        """
        return self._make_request("GET", f"/v1/state/{path}/permissions")

    # Triggers

    def create_trigger(
        self,
        state_path: str,
        trigger_type: str,
        target_function: str,
        key_pattern: Optional[str] = None,
        condition: Optional[Dict[str, Any]] = None,
        target_function_id: Optional[str] = None,
        include_previous: bool = True,
        include_new: bool = True,
        max_invocations_per_minute: int = 60,
        is_active: bool = True
    ) -> Dict[str, Any]:
        """
        Create a trigger for state changes.

        Args:
            state_path: Full path including tenant/state
            trigger_type: Type of trigger (on_set, on_delete, on_change)
            target_function: Name of the function to invoke
            key_pattern: Optional key pattern to match
            condition: Optional condition for trigger
            target_function_id: Optional UUID of target function
            include_previous: Include previous value in trigger payload
            include_new: Include new value in trigger payload
            max_invocations_per_minute: Rate limit for trigger
            is_active: Whether trigger is active

        Returns:
            Created trigger object
        """
        data = {
            "trigger_type": trigger_type,
            "target_function": target_function,
            "include_previous": include_previous,
            "include_new": include_new,
            "max_invocations_per_minute": max_invocations_per_minute,
            "is_active": is_active
        }

        if key_pattern:
            data["key_pattern"] = key_pattern
        if condition:
            data["condition"] = condition
        if target_function_id:
            data["target_function_id"] = target_function_id

        return self._make_request("POST", "/v1/triggers", data=data)

    def get_triggers(
        self,
        state_path: Optional[str] = None,
        is_active: Optional[bool] = None
    ) -> List[Dict[str, Any]]:
        """
        Get triggers, optionally filtered by state path.

        Args:
            state_path: Optional state path to filter by
            is_active: Optional filter by active status

        Returns:
            List of trigger objects
        """
        params = {}
        if state_path:
            params["state_path"] = state_path
        if is_active is not None:
            params["is_active"] = is_active

        return self._make_request("GET", "/v1/triggers", params=params)

    def delete_trigger(self, trigger_id: str) -> None:
        """
        Delete a trigger.

        Args:
            trigger_id: ID of the trigger to delete
        """
        self._make_request("DELETE", f"/v1/triggers/{trigger_id}")


class StateManager:
    """
    State manager for declarative state access in functions.

    Provides a decorator-based interface for managing state within
    FlyPy functions.
    """

    def __init__(
        self,
        api_url: str = DEFAULT_API_URL,
        tenant_id: str = DEFAULT_TENANT_ID,
        api_key: Optional[str] = None
    ):
        """
        Initialize the state manager.

        Args:
            api_url: Base URL for the FunctionFly API
            tenant_id: Tenant ID for multi-tenancy
            api_key: API key for authentication
        """
        self.client = StateClient(api_url, tenant_id, api_key)
        self._state_prefix = tenant_id if tenant_id else "default"

    def state(
        self,
        state_name: str,
        key: Optional[str] = None,
        write: bool = False,
        ttl_days: Optional[int] = None,
        create_if_not_exists: bool = True
    ):
        """
        Decorator for accessing state in a function.

        Args:
            state_name: Name of the state container
            key: Optional key within the state (can be function param)
            write: Whether this is a write operation
            ttl_days: Optional TTL for created state
            create_if_not_exists: Create state if it doesn't exist

        Returns:
            Decorator function
        """
        def decorator(func):
            @wraps(func)
            def wrapper(*args, **kwargs):
                # Build the state path
                state_path = f"{self._state_prefix}/{state_name}"

                # Determine the key (from param or generate from args)
                actual_key = key
                if not actual_key and args:
                    # Use first positional arg as key if not specified
                    actual_key = str(args[0])

                full_path = f"{state_path}/{actual_key}" if actual_key else state_path

                # Try to get the state, create if needed
                if create_if_not_exists:
                    try:
                        self.client.get_state(state_path)
                    except StateNotFoundError:
                        self.client.create_state(
                            state_path,
                            storage_type="durable",
                            ttl_days=ttl_days
                        )

                if write:
                    # Call the function and store result
                    result = func(*args, **kwargs)
                    self.client.set_value(full_path, result)
                    return result
                else:
                    # Get the value and pass to function
                    try:
                        value_data = self.client.get_value(full_path)
                        if value_data and "value" in value_data:
                            return value_data["value"]
                    except StateNotFoundError:
                        pass

                    # If no value, call function with None and let it create
                    return func(*args, **kwargs)

            return wrapper
        return decorator

    def get(self, state_name: str, key: str, default: Any = None) -> Any:
        """
        Get a value from state.

        Args:
            state_name: Name of the state container
            key: Key within the state
            default: Default value if not found

        Returns:
            Value or default
        """
        state_path = f"{self._state_prefix}/{state_name}/{key}"
        try:
            value_data = self.client.get_value(state_path)
            if value_data and "value" in value_data:
                return value_data["value"]
        except StateNotFoundError:
            pass
        return default

    def set(self, state_name: str, key: str, value: Any) -> None:
        """
        Set a value in state.

        Args:
            state_name: Name of the state container
            key: Key within the state
            value: Value to set
        """
        state_path = f"{self._state_prefix}/{state_name}/{key}"
        self.client.set_value(state_path, value)

    def delete(self, state_name: str, key: str) -> None:
        """
        Delete a value from state.

        Args:
            state_name: Name of the state container
            key: Key within the state
        """
        state_path = f"{self._state_prefix}/{state_name}/{key}"
        self.client.delete_value(state_path)

    def snapshot(self, state_name: str, label: Optional[str] = None) -> Dict[str, Any]:
        """
        Create a snapshot of a state container.

        Args:
            state_name: Name of the state container
            label: Optional label for the snapshot

        Returns:
            Snapshot object
        """
        state_path = f"{self._state_prefix}/{state_name}"
        return self.client.create_snapshot(state_path, label)

    def restore(self, state_name: str, snapshot_version: int) -> Dict[str, Any]:
        """
        Restore a state container from snapshot.

        Args:
            state_name: Name of the state container
            snapshot_version: Version to restore

        Returns:
            Restored state
        """
        state_path = f"{self._state_prefix}/{state_name}"
        return self.client.restore_snapshot(state_path, snapshot_version)


# Convenience functions using default client

def get_value(path: str, **kwargs) -> Optional[Dict[str, Any]]:
    """Get a value from state."""
    client = StateClient(**kwargs)
    return client.get_value(path)


def set_value(path: str, value: Dict[str, Any], **kwargs) -> Dict[str, Any]:
    """Set a value in state."""
    client = StateClient(**kwargs)
    return client.set_value(path, value)


def delete_value(path: str, **kwargs) -> None:
    """Delete a value from state."""
    client = StateClient(**kwargs)
    client.delete_value(path)


def get_history(path: str, **kwargs) -> List[Dict[str, Any]]:
    """Get event history for a state path."""
    client = StateClient(**kwargs)
    return client.get_history(path)


def create_snapshot(path: str, label: Optional[str] = None, **kwargs) -> Dict[str, Any]:
    """Create a snapshot."""
    client = StateClient(**kwargs)
    return client.create_snapshot(path, label)


def restore_snapshot(path: str, snapshot_version: int, **kwargs) -> Dict[str, Any]:
    """Restore from a snapshot."""
    client = StateClient(**kwargs)
    return client.restore_snapshot(path, snapshot_version)


# Default client instance for simple usage
_default_client: Optional[StateClient] = None


def get_client() -> StateClient:
    """Get or create the default state client."""
    global _default_client
    if _default_client is None:
        _default_client = StateClient()
    return _default_client


__all__ = [
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
]
