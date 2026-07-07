"""Automatic model rollback when drift is detected.

This module provides automatic rollback of ML models to a previous backup
when significant drift is detected, ensuring service continuity.
"""

import asyncio
import logging
import threading
import time
from dataclasses import dataclass
from datetime import datetime
from enum import StrEnum
from typing import Any, Callable, Dict, List, Optional

from ...config import settings

logger = logging.getLogger(__name__)


class RollbackSeverity(StrEnum):
    """Severity level for rollback triggers."""
    NONE = "none"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


@dataclass
class RollbackEvent:
    """Record of a rollback event."""
    timestamp: datetime
    service: str
    tenant_id: str
    metric_name: str
    severity: RollbackSeverity
    drift_score: float
    backup_version: str
    restored_version: str
    success: bool
    error_message: Optional[str] = None


class AutomaticRollbackManager:
    """Manages automatic model rollback when drift is detected.

    Monitors drift alerts and automatically restores from backup
    when drift exceeds configured thresholds.
    """

    def __init__(self):
        self._lock = threading.Lock()
        self._rollback_callbacks: List[Callable[[str, str, str], bool]] = []
        self._rollback_history: List[RollbackEvent] = []
        self._last_rollback_time: Dict[str, float] = {}
        self._min_rollback_interval_seconds = 3600

        self._auto_rollback_enabled = getattr(
            settings, 'ml_auto_rollback_enabled', False
        )
        self._rollback_threshold = getattr(
            settings, 'ml_auto_rollback_threshold', 0.5
        )
        self._max_rollback_per_day = getattr(
            settings, 'ml_max_rollback_per_day', 3
        )

    def is_enabled(self) -> bool:
        """Check if automatic rollback is enabled."""
        return self._auto_rollback_enabled

    def register_rollback_callback(
        self, callback: Callable[[str, str, str], bool]
    ) -> None:
        """Register a callback to perform model rollback.

        Args:
            callback: Function(service, tenant_id, metric_name) -> bool (success)
        """
        with self._lock:
            self._rollback_callbacks.append(callback)

    def should_rollback(
        self,
        service: str,
        tenant_id: str,
        severity: RollbackSeverity,
        drift_score: float,
    ) -> bool:
        """Determine if automatic rollback should be triggered.

        Args:
            service: ML service name
            tenant_id: Tenant ID
            severity: Drift severity level
            drift_score: Current drift score (0.0-1.0)

        Returns:
            True if rollback should be triggered
        """
        if not self._auto_rollback_enabled:
            return False

        if severity == RollbackSeverity.NONE:
            return False

        if severity in (RollbackSeverity.LOW, RollbackSeverity.MEDIUM):
            return False

        if drift_score < self._rollback_threshold:
            return False

        key = f"{service}:{tenant_id}"
        now = time.time()

        with self._lock:
            if key in self._last_rollback_time:
                last_rollback = self._last_rollback_time[key]
                if now - last_rollback < self._min_rollback_interval_seconds:
                    logger.info(
                        f"Skipping rollback for {key}: too soon since last rollback"
                    )
                    return False

            rollback_today = sum(
                1 for e in self._rollback_history
                if e.service == service
                and e.tenant_id == tenant_id
                and (now - e.timestamp.timestamp()) < 86400
            )

            if rollback_today >= self._max_rollback_per_day:
                logger.info(
                    f"Skipping rollback for {key}: max rollbacks per day reached"
                )
                return False

        return True

    def trigger_rollback(
        self,
        service: str,
        tenant_id: str,
        metric_name: str,
        drift_score: float,
        severity: RollbackSeverity,
        backup_version: str,
    ) -> RollbackEvent:
        """Trigger automatic rollback for a model.

        Args:
            service: ML service name
            tenant_id: Tenant ID
            metric_name: Metric that triggered rollback
            drift_score: Current drift score
            severity: Drift severity level
            backup_version: Version to restore from

        Returns:
            RollbackEvent with results
        """
        key = f"{service}:{tenant_id}"
        now = time.time()
        event = RollbackEvent(
            timestamp=datetime.utcnow(),
            service=service,
            tenant_id=tenant_id,
            metric_name=metric_name,
            severity=severity,
            drift_score=drift_score,
            backup_version=backup_version,
            restored_version="",
            success=False,
        )

        logger.warning(
            f"Triggering automatic rollback for {service}/{tenant_id}: "
            f"drift_score={drift_score:.3f}, severity={severity}, "
            f"backup={backup_version}"
        )

        success = False
        error_message = None

        with self._lock:
            callbacks = list(self._rollback_callbacks)

        for callback in callbacks:
            try:
                success = callback(service, tenant_id, metric_name)
                if success:
                    break
            except Exception as e:
                error_message = str(e)
                logger.error(f"Rollback callback failed: {e}")

        if success:
            event.restored_version = backup_version
            event.success = True
            with self._lock:
                self._last_rollback_time[key] = now
                self._rollback_history.append(event)
            logger.info(
                f"Automatic rollback successful for {service}/{tenant_id}: "
                f"restored to {backup_version}"
            )
        else:
            event.error_message = error_message or "All rollback callbacks failed"
            logger.error(
                f"Automatic rollback failed for {service}/{tenant_id}: "
                f"{event.error_message}"
            )

        return event

    def get_rollback_history(
        self,
        service: Optional[str] = None,
        tenant_id: Optional[str] = None,
        limit: int = 100,
    ) -> List[RollbackEvent]:
        """Get rollback history.

        Args:
            service: Optional filter by service
            tenant_id: Optional filter by tenant
            limit: Maximum number of events to return

        Returns:
            List of RollbackEvent objects
        """
        with self._lock:
            events = list(self._rollback_history)

        if service:
            events = [e for e in events if e.service == service]
        if tenant_id:
            events = [e for e in events if e.tenant_id == tenant_id]

        return sorted(events, key=lambda e: e.timestamp, reverse=True)[:limit]

    def get_stats(self) -> Dict[str, Any]:
        """Get rollback statistics.

        Returns:
            Dict with rollback stats
        """
        with self._lock:
            now = time.time()
            rollbacks_today = sum(
                1 for e in self._rollback_history
                if (now - e.timestamp.timestamp()) < 86400
            )
            successful = sum(1 for e in self._rollback_history if e.success)

            return {
                "auto_rollback_enabled": self._auto_rollback_enabled,
                "rollback_threshold": self._rollback_threshold,
                "total_rollbacks": len(self._rollback_history),
                "successful_rollbacks": successful,
                "failed_rollbacks": len(self._rollback_history) - successful,
                "rollbacks_today": rollbacks_today,
                "max_rollbacks_per_day": self._max_rollback_per_day,
            }


_rollback_manager: Optional[AutomaticRollbackManager] = None


def get_rollback_manager() -> AutomaticRollbackManager:
    """Get the global rollback manager instance."""
    global _rollback_manager
    if _rollback_manager is None:
        _rollback_manager = AutomaticRollbackManager()
    return _rollback_manager


def integrate_drift_with_rollback() -> None:
    """Integrate drift detection with automatic rollback.

    Registers a callback with the drift monitor to trigger
    automatic rollback when significant drift is detected.
    """
    from ..drift_detection import DriftAlert, get_drift_monitor

    rollback_manager = get_rollback_manager()
    drift_monitor = get_drift_monitor()

    def on_drift_detected(alert: DriftAlert) -> None:
        """Handle drift alert and potentially trigger rollback."""
        severity_str = alert.severity.lower()
        try:
            severity = RollbackSeverity(severity_str)
        except ValueError:
            severity = RollbackSeverity.MEDIUM

        if rollback_manager.should_rollback(
            alert.service,
            alert.tenant_id,
            severity,
            alert.drift_score,
        ):
            backup_version = _get_latest_backup_version(
                alert.service, alert.tenant_id
            )
            if backup_version:
                rollback_manager.trigger_rollback(
                    alert.service,
                    alert.tenant_id,
                    alert.metric_name,
                    alert.drift_score,
                    severity,
                    backup_version,
                )
            else:
                logger.warning(
                    f"No backup found for {alert.service}/{alert.tenant_id}, "
                    f"cannot rollback"
                )

    drift_monitor.register_alert_callback(on_drift_detected)
    logger.info("Drift detection integrated with automatic rollback")


def _get_latest_backup_version(service: str, tenant_id: str) -> Optional[str]:
    """Get the latest available backup version for a service.

    Args:
        service: ML service name
        tenant_id: Tenant ID

    Returns:
        Backup version string or None if no backup available
    """
    try:
        from ..persistence import EncryptedModelStore

        store = EncryptedModelStore(namespace=f"{tenant_id}/{service}")
        backups = store.list_backups()

        if backups:
            return backups[0].get("version")

    except Exception as e:
        logger.error(f"Failed to get backup version: {e}")

    return None


async def send_rollback_alert_webhook(event: RollbackEvent) -> bool:
    """Send a rollback alert to webhook URL.

    Args:
        event: The rollback event to report

    Returns:
        True if alert was sent successfully
    """
    import httpx

    webhook_url = getattr(settings, 'ml_rollback_alert_webhook_url', None)
    if not webhook_url:
        return False

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.post(
                webhook_url,
                json={
                    "event": "ml_model_rollback",
                    "service": event.service,
                    "tenant_id": event.tenant_id,
                    "metric_name": event.metric_name,
                    "severity": event.severity.value,
                    "drift_score": event.drift_score,
                    "backup_version": event.backup_version,
                    "restored_version": event.restored_version,
                    "success": event.success,
                    "error_message": event.error_message,
                    "triggered_at": event.timestamp.isoformat(),
                },
            )
            return response.status_code < 400
    except Exception as e:
        logger.error(f"Failed to send rollback alert webhook: {e}")
        return False
