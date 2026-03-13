"""Cache invalidation service.

This module provides intelligent cache invalidation including:
- Tag-based invalidation
- Dependency tracking
- Time-based expiration
- Pattern-based invalidation
"""

import time
import re
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Set
import logging

logger = logging.getLogger(__name__)


class InvalidationType(str, Enum):
    """Types of cache invalidation."""
    EXPLICIT = "explicit"
    TAG_BASED = "tag_based"
    TIME_BASED = "time_based"
    PATTERN_BASED = "pattern_based"
    DEPENDENCY_BASED = "dependency_based"


class InvalidationAction(str, Enum):
    """Actions to take on invalidation."""
    DELETE = "delete"
    EXPIRE = "expire"
    UPDATE = "update"
    TAG = "tag"


@dataclass
class InvalidationRule:
    """A rule for cache invalidation."""
    id: str
    name: str
    invalidation_type: InvalidationType
    pattern: Optional[str] = None  # Regex pattern for keys
    tags: List[str] = field(default_factory=list)  # Tags to match
    ttl: Optional[int] = None  # Time-based expiration
    action: InvalidationAction = InvalidationAction.DELETE
    enabled: bool = True

    def matches(self, key: str, tags: Optional[List[str]] = None) -> bool:
        """Check if a key matches this rule.

        Args:
            key: The cache key
            tags: Tags associated with the key

        Returns:
            True if matches
        """
        if not self.enabled:
            return False

        if self.invalidation_type == InvalidationType.PATTERN_BASED and self.pattern:
            return bool(re.match(self.pattern, key))

        if self.invalidation_type == InvalidationType.TAG_BASED and self.tags:
            if tags is None:
                return False
            return any(tag in tags for tag in self.tags)

        return False


@dataclass
class CacheDependency:
    """A dependency between cache keys."""
    source_key: str
    dependent_keys: Set[str] = field(default_factory=set)
    created_at: float = field(default_factory=time.time)


@dataclass
class InvalidationEvent:
    """An invalidation event."""
    rule_id: Optional[str]
    invalidation_type: InvalidationType
    keys_invalidated: List[str]
    timestamp: datetime = field(default_factory=datetime.utcnow)
    reason: Optional[str] = None


class CacheInvalidator:
    """Manages cache invalidation."""

    def __init__(self):
        """Initialize the cache invalidator."""
        self._rules: Dict[str, InvalidationRule] = {}
        self._dependencies: Dict[str, CacheDependency] = {}
        self._key_tags: Dict[str, Set[str]] = {}  # key -> tags
        self._invalidation_history: List[InvalidationEvent] = []
        self._max_history = 1000
        self._logger = logging.getLogger(__name__)
        self._initialize_default_rules()

    def _initialize_default_rules(self) -> None:
        """Initialize default invalidation rules."""
        # Default TTL rules
        self.add_rule(InvalidationRule(
            id="default_short",
            name="Short TTL (5 minutes)",
            invalidation_type=InvalidationType.TIME_BASED,
            ttl=300,
        ))

        self.add_rule(InvalidationRule(
            id="default_medium",
            name="Medium TTL (1 hour)",
            invalidation_type=InvalidationType.TIME_BASED,
            ttl=3600,
        ))

        self.add_rule(InvalidationRule(
            id="default_long",
            name="Long TTL (24 hours)",
            invalidation_type=InvalidationType.TIME_BASED,
            ttl=86400,
        ))

    def add_rule(self, rule: InvalidationRule) -> None:
        """Add an invalidation rule.

        Args:
            rule: The rule to add
        """
        self._rules[rule.id] = rule
        self._logger.info(f"Added invalidation rule: {rule.name}")

    def remove_rule(self, rule_id: str) -> bool:
        """Remove an invalidation rule.

        Args:
            rule_id: The rule ID to remove

        Returns:
            True if removed, False if not found
        """
        if rule_id in self._rules:
            del self._rules[rule_id]
            return True
        return False

    def get_rule(self, rule_id: str) -> Optional[InvalidationRule]:
        """Get an invalidation rule.

        Args:
            rule_id: The rule ID

        Returns:
            InvalidationRule or None
        """
        return self._rules.get(rule_id)

    def list_rules(self, enabled_only: bool = True) -> List[InvalidationRule]:
        """List all invalidation rules.

        Args:
            enabled_only: Only return enabled rules

        Returns:
            List of InvalidationRule
        """
        rules = list(self._rules.values())
        if enabled_only:
            rules = [r for r in rules if r.enabled]
        return rules

    def register_dependency(
        self,
        source_key: str,
        dependent_keys: List[str],
    ) -> None:
        """Register a dependency between cache keys.

        When source_key is invalidated, all dependent_keys will also be invalidated.

        Args:
            source_key: The source key
            dependent_keys: Keys that depend on source_key
        """
        if source_key not in self._dependencies:
            self._dependencies[source_key] = CacheDependency(
                source_key=source_key,
            )

        self._dependencies[source_key].dependent_keys.update(dependent_keys)

        # Also register reverse dependencies
        for dep_key in dependent_keys:
            if dep_key not in self._dependencies:
                self._dependencies[dep_key] = CacheDependency(source_key=dep_key)
            self._dependencies[dep_key].dependent_keys.add(source_key)

    def register_tags(self, key: str, tags: List[str]) -> None:
        """Register tags for a cache key.

        Args:
            key: The cache key
            tags: Tags to associate with the key
        """
        if key not in self._key_tags:
            self._key_tags[key] = set()
        self._key_tags[key].update(tags)

    def get_keys_for_tags(self, tags: List[str]) -> Set[str]:
        """Get all keys associated with any of the given tags.

        Args:
            tags: Tags to search for

        Returns:
            Set of cache keys
        """
        keys = set()
        for key, key_tags in self._key_tags.items():
            if any(tag in key_tags for tag in tags):
                keys.add(key)
        return keys

    def get_dependent_keys(self, key: str) -> Set[str]:
        """Get all keys that depend on the given key.

        Args:
            key: The cache key

        Returns:
            Set of dependent keys
        """
        if key not in self._dependencies:
            return set()
        return self._dependencies[key].dependent_keys.copy()

    def invalidate_by_key(
        self,
        key: str,
        delete_callback: Callable[[str], bool],
    ) -> List[str]:
        """Invalidate a specific key and its dependencies.

        Args:
            key: The cache key to invalidate
            delete_callback: Callback to actually delete the key

        Returns:
            List of invalidated keys
        """
        invalidated = [key]

        # Delete the key
        try:
            delete_callback(key)
        except Exception as e:
            self._logger.error(f"Error deleting key {key}: {e}")

        # Get and invalidate dependencies
        dependent_keys = self.get_dependent_keys(key)
        for dep_key in dependent_keys:
            try:
                if delete_callback(dep_key):
                    invalidated.append(dep_key)
            except Exception as e:
                self._logger.error(f"Error deleting dependent key {dep_key}: {e}")

        # Record event
        self._record_invalidation(
            rule_id=None,
            invalidation_type=InvalidationType.EXPLICIT,
            keys_invalidated=invalidated,
            reason=f"Explicit invalidation of {key}",
        )

        return invalidated

    def invalidate_by_tags(
        self,
        tags: List[str],
        delete_callback: Callable[[str], bool],
    ) -> List[str]:
        """Invalidate all keys with the given tags.

        Args:
            tags: Tags to invalidate
            delete_callback: Callback to delete keys

        Returns:
            List of invalidated keys
        """
        keys = self.get_keys_for_tags(tags)
        invalidated = []

        for key in keys:
            try:
                if delete_callback(key):
                    invalidated.append(key)
            except Exception as e:
                self._logger.error(f"Error deleting key {key}: {e}")

        # Record event
        self._record_invalidation(
            rule_id=None,
            invalidation_type=InvalidationType.TAG_BASED,
            keys_invalidated=invalidated,
            reason=f"Tag-based invalidation: {tags}",
        )

        return invalidated

    def invalidate_by_pattern(
        self,
        pattern: str,
        delete_callback: Callable[[str], bool],
    ) -> List[str]:
        """Invalidate all keys matching a pattern.

        Args:
            pattern: Regex pattern for keys
            delete_callback: Callback to delete keys

        Returns:
            List of invalidated keys
        """
        regex = re.compile(pattern)

        # Find matching keys
        matching_keys = [
            key for key in self._key_tags.keys()
            if regex.match(key)
        ]

        invalidated = []
        for key in matching_keys:
            try:
                if delete_callback(key):
                    invalidated.append(key)
            except Exception as e:
                self._logger.error(f"Error deleting key {key}: {e}")

        # Record event
        self._record_invalidation(
            rule_id=None,
            invalidation_type=InvalidationType.PATTERN_BASED,
            keys_invalidated=invalidated,
            reason=f"Pattern-based invalidation: {pattern}",
        )

        return invalidated

    def invalidate_all(self, clear_callback: Callable[[], None]) -> int:
        """Invalidate all cache entries.

        Args:
            clear_callback: Callback to clear the cache

        Returns:
            Number of entries cleared (approximate)
        """
        count = len(self._key_tags)

        try:
            clear_callback()
        except Exception as e:
            self._logger.error(f"Error clearing cache: {e}")

        # Record event
        self._record_invalidation(
            rule_id=None,
            invalidation_type=InvalidationType.EXPLICIT,
            keys_invalidated=list(self._key_tags.keys()),
            reason="Cache clear all",
        )

        # Clear internal state
        self._key_tags.clear()
        self._dependencies.clear()

        return count

    def _record_invalidation(
        self,
        rule_id: Optional[str],
        invalidation_type: InvalidationType,
        keys_invalidated: List[str],
        reason: Optional[str] = None,
    ) -> None:
        """Record an invalidation event.

        Args:
            rule_id: The rule that triggered invalidation
            invalidation_type: Type of invalidation
            keys_invalidated: Keys that were invalidated
            reason: Reason for invalidation
        """
        event = InvalidationEvent(
            rule_id=rule_id,
            invalidation_type=invalidation_type,
            keys_invalidated=keys_invalidated,
            reason=reason,
        )

        self._invalidation_history.append(event)

        # Trim history if needed
        if len(self._invalidation_history) > self._max_history:
            self._invalidation_history = self._invalidation_history[-self._max_history:]

    def get_invalidation_history(
        self,
        limit: int = 100,
    ) -> List[InvalidationEvent]:
        """Get invalidation history.

        Args:
            limit: Maximum number of events to return

        Returns:
            List of InvalidationEvent
        """
        return self._invalidation_history[-limit:]

    def get_stats(self) -> Dict[str, Any]:
        """Get invalidator statistics.

        Returns:
            Dictionary with statistics
        """
        total_keys = len(self._key_tags)
        total_dependencies = len(self._dependencies)

        # Count keys per tag
        tag_counts: Dict[str, int] = {}
        for tags in self._key_tags.values():
            for tag in tags:
                tag_counts[tag] = tag_counts.get(tag, 0) + 1

        return {
            "total_keys": total_keys,
            "total_dependencies": total_dependencies,
            "total_rules": len(self._rules),
            "enabled_rules": len([r for r in self._rules.values() if r.enabled]),
            "tag_counts": tag_counts,
            "invalidation_history_count": len(self._invalidation_history),
        }


# Global invalidator instance
_invalidator: Optional[CacheInvalidator] = None


def get_cache_invalidator() -> CacheInvalidator:
    """Get the global cache invalidator instance.

    Returns:
        CacheInvalidator instance
    """
    global _invalidator
    if _invalidator is None:
        _invalidator = CacheInvalidator()

    return _invalidator
