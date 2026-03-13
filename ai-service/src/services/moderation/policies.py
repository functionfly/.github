"""Moderation policy definition and management.

This module handles the creation, storage, and retrieval of moderation
policies that define how different types of content are handled.
"""

from datetime import datetime
from enum import Enum
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
import uuid


class ModerationCategory(str, Enum):
    """Categories of content that can be flagged."""
    PII = "pii"
    SECRETS = "secrets"
    TOXIC = "toxic"
    MALWARE = "malware"
    HATE_SPEECH = "hate_speech"
    VIOLENCE = "violence"
    SEXUAL = "sexual"
    SPAM = "spam"
    CUSTOM = "custom"


class PolicyAction(str, Enum):
    """Actions to take when a policy is violated."""
    ALLOW = "allow"      # Allow but log
    FLAG = "flag"        # Allow but flag for review
    BLOCK = "block"      # Block the content
    QUARANTINE = "quarantine"  # Quarantine for admin review


class PolicyRule(BaseModel):
    """A single rule within a moderation policy."""
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    category: ModerationCategory
    action: PolicyAction
    severity_threshold: str = "low"  # low, medium, high, critical
    enabled: bool = True
    custom_pattern: Optional[str] = None  # Optional custom regex pattern


class ModerationPolicy(BaseModel):
    """A moderation policy that defines how content is handled."""
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    name: str
    description: Optional[str] = None
    tenant_id: Optional[str] = None  # None means global policy
    rules: List[PolicyRule] = Field(default_factory=list)
    is_default: bool = False
    is_active: bool = True
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    created_by: Optional[str] = None

    def get_rule(self, category: ModerationCategory) -> Optional[PolicyRule]:
        """Get a rule for a specific category.

        Args:
            category: The category to find a rule for

        Returns:
            PolicyRule if found, None otherwise
        """
        for rule in self.rules:
            if rule.category == category and rule.enabled:
                return rule
        return None

    def add_rule(self, rule: PolicyRule) -> None:
        """Add a rule to the policy.

        Args:
            rule: The rule to add
        """
        # Remove existing rule for the same category
        self.rules = [r for r in self.rules if r.category != rule.category]
        self.rules.append(rule)
        self.updated_at = datetime.utcnow()


class PolicyManager:
    """Manages moderation policies."""

    def __init__(self):
        """Initialize the policy manager."""
        self._policies: Dict[str, ModerationPolicy] = {}
        self._tenant_policies: Dict[str, str] = {}  # tenant_id -> policy_id
        self._default_policy_id: Optional[str] = None
        self._initialize_defaults()

    def _initialize_defaults(self) -> None:
        """Initialize default policies."""
        # Create default global policy
        default_policy = ModerationPolicy(
            name="Default Policy",
            description="Default moderation policy for all content",
            is_default=True,
            is_active=True,
            rules=[
                # Block secrets and API keys
                PolicyRule(
                    category=ModerationCategory.SECRETS,
                    action=PolicyAction.BLOCK,
                    severity_threshold="low",
                ),
                # Flag PII for review
                PolicyRule(
                    category=ModerationCategory.PII,
                    action=PolicyAction.FLAG,
                    severity_threshold="low",
                ),
                # Block known malware patterns
                PolicyRule(
                    category=ModerationCategory.MALWARE,
                    action=PolicyAction.BLOCK,
                    severity_threshold="low",
                ),
                # Flag toxic content
                PolicyRule(
                    category=ModerationCategory.TOXIC,
                    action=PolicyAction.FLAG,
                    severity_threshold="medium",
                ),
                # Block hate speech
                PolicyRule(
                    category=ModerationCategory.HATE_SPEECH,
                    action=PolicyAction.BLOCK,
                    severity_threshold="medium",
                ),
                # Block violence
                PolicyRule(
                    category=ModerationCategory.VIOLENCE,
                    action=PolicyAction.BLOCK,
                    severity_threshold="medium",
                ),
                # Flag sexual content
                PolicyRule(
                    category=ModerationCategory.SEXUAL,
                    action=PolicyAction.FLAG,
                    severity_threshold="high",
                ),
                # Flag spam
                PolicyRule(
                    category=ModerationCategory.SPAM,
                    action=PolicyAction.FLAG,
                    severity_threshold="medium",
                ),
            ],
        )

        self._policies[default_policy.id] = default_policy
        self._default_policy_id = default_policy.id

    def get_policy(self, tenant_id: Optional[str] = None) -> Optional[ModerationPolicy]:
        """Get a policy for a tenant.

        Args:
            tenant_id: The tenant ID to get a policy for

        Returns:
            ModerationPolicy if found, None otherwise
        """
        if tenant_id and tenant_id in self._tenant_policies:
            policy_id = self._tenant_policies[tenant_id]
            return self._policies.get(policy_id)

        # Return default policy
        if self._default_policy_id:
            return self._policies.get(self._default_policy_id)

        return None

    def get_policy_by_id(self, policy_id: str) -> Optional[ModerationPolicy]:
        """Get a policy by its ID.

        Args:
            policy_id: The policy ID

        Returns:
            ModerationPolicy if found, None otherwise
        """
        return self._policies.get(policy_id)

    def list_policies(
        self,
        tenant_id: Optional[str] = None,
        include_inactive: bool = False,
    ) -> List[ModerationPolicy]:
        """List all policies.

        Args:
            tenant_id: Filter by tenant ID
            include_inactive: Include inactive policies

        Returns:
            List of ModerationPolicy
        """
        policies = list(self._policies.values())

        if tenant_id:
            policies = [p for p in policies if p.tenant_id == tenant_id]

        if not include_inactive:
            policies = [p for p in policies if p.is_active]

        return sorted(policies, key=lambda p: p.created_at, reverse=True)

    def create_policy(
        self,
        name: str,
        description: Optional[str] = None,
        tenant_id: Optional[str] = None,
        rules: Optional[List[PolicyRule]] = None,
        created_by: Optional[str] = None,
    ) -> ModerationPolicy:
        """Create a new policy.

        Args:
            name: Policy name
            description: Policy description
            tenant_id: Tenant ID (None for global)
            rules: List of policy rules
            created_by: User who created the policy

        Returns:
            The created ModerationPolicy
        """
        policy = ModerationPolicy(
            name=name,
            description=description,
            tenant_id=tenant_id,
            rules=rules or [],
            created_by=created_by,
        )

        self._policies[policy.id] = policy

        # Set as tenant's default if specified
        if tenant_id:
            self._tenant_policies[tenant_id] = policy.id

        return policy

    def update_policy(
        self,
        policy_id: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        rules: Optional[List[PolicyRule]] = None,
        is_active: Optional[bool] = None,
    ) -> Optional[ModerationPolicy]:
        """Update an existing policy.

        Args:
            policy_id: The policy ID to update
            name: New name
            description: New description
            rules: New rules
            is_active: Active status

        Returns:
            Updated ModerationPolicy or None if not found
        """
        policy = self._policies.get(policy_id)
        if policy is None:
            return None

        if name is not None:
            policy.name = name
        if description is not None:
            policy.description = description
        if rules is not None:
            policy.rules = rules
        if is_active is not None:
            policy.is_active = is_active

        policy.updated_at = datetime.utcnow()

        return policy

    def delete_policy(self, policy_id: str) -> bool:
        """Delete a policy.

        Args:
            policy_id: The policy ID to delete

        Returns:
            True if deleted, False if not found or if it's the default
        """
        policy = self._policies.get(policy_id)
        if policy is None:
            return False

        if policy.is_default:
            return False  # Cannot delete default policy

        # Remove tenant mapping if exists
        if policy.tenant_id and policy.tenant_id in self._tenant_policies:
            del self._tenant_policies[policy.tenant_id]

        del self._policies[policy_id]
        return True

    def set_default_policy(self, policy_id: str) -> bool:
        """Set a policy as the default.

        Args:
            policy_id: The policy ID to set as default

        Returns:
            True if set, False if not found
        """
        policy = self._policies.get(policy_id)
        if policy is None:
            return False

        # Unset current default
        if self._default_policy_id:
            current_default = self._policies.get(self._default_policy_id)
            if current_default:
                current_default.is_default = False

        # Set new default
        policy.is_default = True
        self._default_policy_id = policy_id

        return True


# Global policy manager instance
_policy_manager: Optional[PolicyManager] = None


def get_policy_manager() -> PolicyManager:
    """Get the global policy manager instance.

    Returns:
        PolicyManager instance
    """
    global _policy_manager
    if _policy_manager is None:
        _policy_manager = PolicyManager()

    return _policy_manager
