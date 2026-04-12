"""Security alerting for FlyMind AI Service.

This module provides security alerting capabilities for critical events
like auth failures, PII violations, cost threshold breaches, etc.
"""

import logging
import json
from dataclasses import dataclass
from datetime import datetime
from typing import Optional, Dict, Any, List
from enum import Enum

logger = logging.getLogger(__name__)


class AlertSeverity(str, Enum):
    """Alert severity levels."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class AlertType(str, Enum):
    """Types of security alerts."""
    AUTH_FAILURE = "auth_failure"
    PII_VIOLATION = "pii_violation"
    COST_THRESHOLD = "cost_threshold"
    RATE_LIMIT_EXCEEDED = "rate_limit_exceeded"
    SUSPICIOUS_ACTIVITY = "suspicious_activity"
    EMBEDDING_BLOCKED = "embedding_blocked"
    RAG_CONTENT_BLOCKED = "rag_content_blocked"


@dataclass
class SecurityAlert:
    """Security alert data structure."""
    alert_type: str
    severity: str
    tenant_id: str
    message: str
    timestamp: datetime
    metadata: Dict[str, Any]
    acknowledged: bool = False
    acknowledged_at: Optional[datetime] = None
    acknowledged_by: Optional[str] = None


class SecurityAlerter:
    """Manages security alerts for the AI service."""
    
    def __init__(self):
        self._logger = logging.getLogger("flymind.security.alerts")
        self._alert_handlers: List[callable] = []
        self._alert_thresholds: Dict[str, Dict[str, Any]] = {
            AlertType.AUTH_FAILURE: {"count": 0, "threshold": 10, "window_minutes": 5},
            AlertType.PII_VIOLATION: {"count": 0, "threshold": 5, "window_minutes": 5},
            AlertType.COST_THRESHOLD: {"count": 0, "threshold": 1, "window_minutes": 60},
        }
        self._alert_history: List[SecurityAlert] = []
        self._max_history = 1000
    
    def register_handler(self, handler: callable) -> None:
        """Register an alert handler.
        
        Args:
            handler: Function that takes a SecurityAlert and sends it
                    (e.g., to Slack, email, PagerDuty)
        """
        self._alert_handlers.append(handler)
    
    async def send_alert(
        self,
        alert_type: AlertType,
        severity: AlertSeverity,
        tenant_id: str,
        message: str,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> SecurityAlert:
        """Send a security alert.
        
        Args:
            alert_type: Type of alert
            severity: Severity level
            tenant_id: Tenant ID
            message: Alert message
            metadata: Additional metadata
            
        Returns:
            Created alert
        """
        alert = SecurityAlert(
            alert_type=alert_type.value,
            severity=severity.value,
            tenant_id=tenant_id,
            message=message,
            timestamp=datetime.utcnow(),
            metadata=metadata or {},
        )
        
        # Log to structured logger
        self._logger.critical(
            f"Security Alert: {alert_type.value}",
            extra={
                "alert_type": alert_type.value,
                "severity": severity.value,
                "tenant_id": tenant_id,
                "message": message,
                "metadata": metadata,
            }
        )
        
        # Store in history
        self._alert_history.append(alert)
        if len(self._alert_history) > self._max_history:
            self._alert_history = self._alert_history[-self._max_history:]
        
        # Send to all registered handlers
        for handler in self._alert_handlers:
            try:
                await handler(alert)
            except Exception as e:
                logger.error(f"Alert handler failed: {e}")
        
        return alert
    
    async def check_and_alert(
        self,
        alert_type: AlertType,
        tenant_id: str,
        message: str,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> Optional[SecurityAlert]:
        """Check threshold and send alert if exceeded.
        
        This implements rate limiting for alerts to prevent alert fatigue.
        
        Args:
            alert_type: Type of alert
            tenant_id: Tenant ID
            message: Alert message
            metadata: Additional metadata
            
        Returns:
            Alert if sent, None if suppressed
        """
        threshold_config = self._alert_thresholds.get(alert_type)
        if not threshold_config:
            # No threshold, always alert
            return await self.send_alert(
                alert_type,
                AlertSeverity.HIGH,
                tenant_id,
                message,
                metadata,
            )
        
        # Increment counter
        threshold_config["count"] += 1
        
        # Check if threshold exceeded
        if threshold_config["count"] >= threshold_config["threshold"]:
            # Send alert and reset counter
            severity = AlertSeverity.CRITICAL if threshold_config["count"] >= threshold_config["threshold"] * 2 else AlertSeverity.HIGH
            alert = await self.send_alert(
                alert_type,
                severity,
                tenant_id,
                f"Threshold exceeded: {message}",
                {
                    **(metadata or {}),
                    "threshold": threshold_config["threshold"],
                    "actual_count": threshold_config["count"],
                    "window_minutes": threshold_config["window_minutes"],
                },
            )
            threshold_config["count"] = 0
            return alert
        
        return None
    
    def acknowledge_alert(
        self,
        alert_index: int,
        acknowledged_by: str,
    ) -> bool:
        """Acknowledge an alert.
        
        Args:
            alert_index: Index in alert history
            acknowledged_by: Who acknowledged it
            
        Returns:
            True if acknowledged, False if not found
        """
        if 0 <= alert_index < len(self._alert_history):
            alert = self._alert_history[alert_index]
            alert.acknowledged = True
            alert.acknowledged_at = datetime.utcnow()
            alert.acknowledged_by = acknowledged_by
            return True
        return False
    
    def get_alert_history(
        self,
        tenant_id: Optional[str] = None,
        severity: Optional[str] = None,
        acknowledged: Optional[bool] = None,
        limit: int = 100,
    ) -> List[SecurityAlert]:
        """Get alert history with optional filtering.
        
        Args:
            tenant_id: Filter by tenant
            severity: Filter by severity
            acknowledged: Filter by acknowledged status
            limit: Maximum results
            
        Returns:
            List of alerts
        """
        alerts = self._alert_history
        
        if tenant_id:
            alerts = [a for a in alerts if a.tenant_id == tenant_id]
        if severity:
            alerts = [a for a in alerts if a.severity == severity]
        if acknowledged is not None:
            alerts = [a for a in alerts if a.acknowledged == acknowledged]
        
        return alerts[-limit:]
    
    def get_stats(self) -> Dict[str, Any]:
        """Get alerter statistics.
        
        Returns:
            Dictionary with stats
        """
        total_alerts = len(self._alert_history)
        unacknowledged = len([a for a in self._alert_history if not a.acknowledged])
        by_severity: Dict[str, int] = {}
        for alert in self._alert_history:
            by_severity[alert.severity] = by_severity.get(alert.severity, 0) + 1
        
        return {
            "total_alerts": total_alerts,
            "unacknowledged": unacknowledged,
            "by_severity": by_severity,
            "handlers_registered": len(self._alert_handlers),
        }


# Global alerter instance
_alerter: Optional[SecurityAlerter] = None


def get_security_alerter() -> SecurityAlerter:
    """Get the global security alerter instance.
    
    Returns:
        SecurityAlerter instance
    """
    global _alerter
    if _alerter is None:
        _alerter = SecurityAlerter()
    return _alerter


# Example alert handlers
async def console_alert_handler(alert: SecurityAlert) -> None:
    """Simple console alert handler for development."""
    print(f"\n🔒 SECURITY ALERT [{alert.severity.upper()}]")
    print(f"Type: {alert.alert_type}")
    print(f"Tenant: {alert.tenant_id}")
    print(f"Message: {alert.message}")
    print(f"Time: {alert.timestamp.isoformat()}")
    if alert.metadata:
        print(f"Metadata: {json.dumps(alert.metadata, indent=2)}")
    print("-" * 50)


def register_default_handlers() -> None:
    """Register default alert handlers."""
    alerter = get_security_alerter()
    alerter.register_handler(console_alert_handler)
