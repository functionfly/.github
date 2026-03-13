"""Moderation result models and service.

This module contains the data models for moderation results and the main
moderation service that orchestrates content scanning.
"""

from datetime import datetime
from enum import Enum
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
import uuid


class ModerationDecision(str, Enum):
    """Decision made by the moderation system."""
    ALLOW = "allow"
    BLOCK = "block"
    FLAG = "flag"
    QUARANTINE = "quarantine"


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


class Violation(BaseModel):
    """A single violation found in content."""
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    category: ModerationCategory
    severity: str = "medium"  # low, medium, high, critical
    message: str
    matched_pattern: Optional[str] = None
    location_start: Optional[int] = None
    location_end: Optional[int] = None
    confidence: float = Field(ge=0.0, le=1.0, default=0.8)


class ScanRequest(BaseModel):
    """Request to scan content."""
    content: str
    content_type: str = "text"  # text, json, code
    tenant_id: Optional[str] = None
    user_id: Optional[str] = None
    context: Dict[str, Any] = Field(default_factory=dict)


class ModerationResult(BaseModel):
    """Result of content moderation."""
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    decision: ModerationDecision
    violations: List[Violation] = Field(default_factory=list)
    scanned_at: datetime = Field(default_factory=datetime.utcnow)
    scan_duration_ms: float = 0.0
    content_hash: Optional[str] = None
    tenant_id: Optional[str] = None
    user_id: Optional[str] = None
    metadata: Dict[str, Any] = Field(default_factory=dict)

    @property
    def is_allowed(self) -> bool:
        """Check if content is allowed."""
        return self.decision in (ModerationDecision.ALLOW, ModerationDecision.FLAG)

    @property
    def has_violations(self) -> bool:
        """Check if any violations were found."""
        return len(self.violations) > 0

    @property
    def severity_score(self) -> float:
        """Calculate overall severity score."""
        if not self.violations:
            return 0.0
        severity_weights = {
            "low": 0.25,
            "medium": 0.5,
            "high": 0.75,
            "critical": 1.0,
        }
        return max(severity_weights.get(v.severity, 0.5) for v in self.violations)


class ModerationService:
    """Main moderation service that orchestrates content scanning."""

    def __init__(self, scanner, policy_manager):
        """Initialize the moderation service.

        Args:
            scanner: Content scanner instance
            policy_manager: Policy manager instance
        """
        self._scanner = scanner
        self._policy_manager = policy_manager

    async def scan(
        self,
        content: str,
        content_type: str = "text",
        tenant_id: Optional[str] = None,
        user_id: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None,
    ) -> ModerationResult:
        """Scan content for policy violations.

        Args:
            content: Content to scan
            content_type: Type of content (text, json, code)
            tenant_id: Tenant ID for policy lookup
            user_id: User ID for audit logging
            context: Additional context for scanning

        Returns:
            ModerationResult with decision and violations
        """
        import time
        start_time = time.time()

        # Get applicable policy
        policy = self._policy_manager.get_policy(tenant_id)

        # Scan content
        violations = await self._scanner.scan(content, content_type)

        # Apply policy rules to determine decision
        decision = self._apply_policy(violations, policy)

        # Calculate scan duration
        scan_duration_ms = (time.time() - start_time) * 1000

        # Build result
        result = ModerationResult(
            decision=decision,
            violations=violations,
            scan_duration_ms=scan_duration_ms,
            tenant_id=tenant_id,
            user_id=user_id,
            metadata={
                "content_type": content_type,
                "policy_id": policy.id if policy else None,
                "context": context or {},
            },
        )

        return result

    def _apply_policy(
        self,
        violations: List[Violation],
        policy,
    ) -> ModerationDecision:
        """Apply policy rules to determine the decision.

        Args:
            violations: List of violations found
            policy: Applied policy

        Returns:
            ModerationDecision
        """
        if not violations:
            return ModerationDecision.ALLOW

        if policy is None:
            # Default: block on critical violations
            if any(v.severity == "critical" for v in violations):
                return ModerationDecision.BLOCK
            return ModerationDecision.FLAG

        # Check each violation against policy rules
        block_categories = set()
        quarantine_categories = set()
        flag_categories = set()

        for rule in policy.rules:
            if rule.action == PolicyAction.BLOCK:
                block_categories.add(rule.category)
            elif rule.action == PolicyAction.QUARANTINE:
                quarantine_categories.add(rule.category)
            elif rule.action == PolicyAction.FLAG:
                flag_categories.add(rule.category)

        # Determine decision based on violations
        for violation in violations:
            if violation.category in block_categories:
                return ModerationDecision.BLOCK
            if violation.category in quarantine_categories:
                return ModerationDecision.QUARANTINE

        # If only flag categories are violated
        if any(v.category in flag_categories for v in violations):
            return ModerationDecision.FLAG

        # Default to flag for unknown violations
        return ModerationDecision.FLAG


# Global service instance
_moderation_service: Optional[ModerationService] = None


def get_moderation_service() -> ModerationService:
    """Get the global moderation service instance.

    Returns:
        ModerationService instance
    """
    global _moderation_service
    if _moderation_service is None:
        from .scanner import get_content_scanner
        from .policies import get_policy_manager

        scanner = get_content_scanner()
        policy_manager = get_policy_manager()
        _moderation_service = ModerationService(scanner, policy_manager)

    return _moderation_service
