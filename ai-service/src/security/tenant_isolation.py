"""Tenant isolation for FlyMind AI Service.

This module provides multi-tenant isolation to ensure tenants cannot
access each other's resources.
"""

import contextvars
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Set
import threading


# Context variable for current tenant
_current_tenant: contextvars.ContextVar[Optional["TenantContext"]] = contextvars.ContextVar(
    "current_tenant",
    default=None
)


@dataclass
class TenantContext:
    """Context for the current tenant."""
    tenant_id: str
    user_id: Optional[str] = None
    api_key_id: Optional[str] = None
    scopes: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)

    def has_scope(self, scope: str) -> bool:
        """Check if the context has a specific scope."""
        return scope in self.scopes or "admin" in self.scopes or "full" in self.scopes

    def can_access_tenant(self, target_tenant_id: str) -> bool:
        """Check if the context can access a target tenant."""
        # Admin can access any tenant
        if "admin" in self.scopes:
            return True

        # Otherwise, must match
        return self.tenant_id == target_tenant_id


class TenantIsolation:
    """Manages tenant isolation for the service."""

    def __init__(self):
        """Initialize the tenant isolation manager."""
        self._lock = threading.Lock()

        # Tenant configurations
        self._tenant_configs: Dict[str, Dict[str, Any]] = {}

        # Allowed resources per tenant
        self._tenant_resources: Dict[str, Set[str]] = {}

        # Initialize default tenant
        self._init_default_tenant()

    def _init_default_tenant(self) -> None:
        """Initialize the default tenant."""
        self._tenant_configs["default"] = {
            "name": "Default Tenant",
            "rate_limit": 60,
            "cost_limit": 100.0,
            "enabled": True,
        }
        self._tenant_resources["default"] = set()

    def register_tenant(
        self,
        tenant_id: str,
        name: str,
        rate_limit: int = 60,
        cost_limit: float = 100.0,
    ) -> None:
        """Register a new tenant.

        Args:
            tenant_id: Tenant ID
            name: Tenant name
            rate_limit: Requests per minute
            cost_limit: Daily cost limit in USD
        """
        with self._lock:
            self._tenant_configs[tenant_id] = {
                "name": name,
                "rate_limit": rate_limit,
                "cost_limit": cost_limit,
                "enabled": True,
            }
            self._tenant_resources[tenant_id] = set()

    def get_tenant_config(self, tenant_id: str) -> Optional[Dict[str, Any]]:
        """Get tenant configuration.

        Args:
            tenant_id: Tenant ID

        Returns:
            Tenant configuration or None
        """
        return self._tenant_configs.get(tenant_id)

    def is_tenant_enabled(self, tenant_id: str) -> bool:
        """Check if a tenant is enabled.

        Args:
            tenant_id: Tenant ID

        Returns:
            True if enabled
        """
        config = self.get_tenant_config(tenant_id)
        if not config:
            return False
        return config.get("enabled", False)

    def set_context(self, context: TenantContext) -> None:
        """Set the current tenant context.

        Args:
            context: Tenant context
        """
        _current_tenant.set(context)

    def get_context(self) -> Optional[TenantContext]:
        """Get the current tenant context.

        Returns:
            TenantContext or None
        """
        return _current_tenant.get()

    def clear_context(self) -> None:
        """Clear the current tenant context."""
        _current_tenant.set(None)

    def check_resource_access(
        self,
        resource_id: str,
        resource_type: str = "function",
    ) -> bool:
        """Check if the current tenant can access a resource.

        Args:
            resource_id: Resource ID
            resource_type: Type of resource

        Returns:
            True if access allowed
        """
        context = self.get_context()
        if not context:
            return False

        # Admin has full access
        if "admin" in context.scopes:
            return True

        # Build resource key
        resource_key = f"{resource_type}:{resource_id}"

        # Check if tenant has this resource
        if resource_id in self._tenant_resources.get(context.tenant_id, set()):
            return True

        # Check if resource is in tenant's namespace
        if resource_id.startswith(f"{context.tenant_id}:"):
            return True

        return False

    def add_resource(
        self,
        tenant_id: str,
        resource_id: str,
    ) -> None:
        """Add a resource to a tenant.

        Args:
            tenant_id: Tenant ID
            resource_id: Resource ID
        """
        with self._lock:
            if tenant_id not in self._tenant_resources:
                self._tenant_resources[tenant_id] = set()
            self._tenant_resources[tenant_id].add(resource_id)

    def remove_resource(
        self,
        tenant_id: str,
        resource_id: str,
    ) -> None:
        """Remove a resource from a tenant.

        Args:
            tenant_id: Tenant ID
            resource_id: Resource ID
        """
        with self._lock:
            if tenant_id in self._tenant_resources:
                self._tenant_resources[tenant_id].discard(resource_id)

    def list_tenant_resources(
        self,
        tenant_id: str,
    ) -> List[str]:
        """List all resources for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            List of resource IDs
        """
        with self._lock:
            return list(self._tenant_resources.get(tenant_id, set()))

    def filter_by_tenant(
        self,
        items: List[Dict[str, Any]],
        id_field: str = "id",
    ) -> List[Dict[str, Any]]:
        """Filter items by the current tenant.

        Args:
            items: List of items to filter
            id_field: Field name for the ID

        Returns:
            Filtered list
        """
        context = self.get_context()
        if not context:
            return []

        # Admin sees all
        if "admin" in context.scopes:
            return items

        return [
            item for item in items
            if item.get(id_field, "").startswith(context.tenant_id)
        ]


# Global isolation manager
_tenant_isolation: Optional[TenantIsolation] = None


def get_tenant_isolation() -> TenantIsolation:
    """Get the global tenant isolation manager.

    Returns:
        TenantIsolation instance
    """
    global _tenant_isolation
    if _tenant_isolation is None:
        _tenant_isolation = TenantIsolation()

    return _tenant_isolation
