"""Alerting service for anomaly detection.

Generates and manages alerts for detected anomalies.
"""

import logging
from datetime import datetime
from typing import Optional

from .detector import AnomalyDetector, get_anomaly_detector
from .models import AnomalyRecord

logger = logging.getLogger(__name__)


class AlertingService:
    """Service for generating and managing alerts.

    Handles alert generation, notification, and lifecycle management.
    """

    def __init__(self):
        self._detector = get_anomaly_detector()
        self._alert_handlers = []

    def register_handler(self, handler) -> None:
        """Register an alert handler.

        Args:
            handler: Callback function that handles alerts
        """
        self._alert_handlers.append(handler)
        logger.info(f"Registered alert handler: {handler.__name__}")

    async def handle_anomaly(self, anomaly: AnomalyRecord) -> None:
        """Handle a detected anomaly by generating an alert.

        Args:
            anomaly: The detected anomaly
        """
        # Store the anomaly
        await self._detector.store_anomaly(anomaly)

        # Notify registered handlers
        for handler in self._alert_handlers:
            try:
                await handler(anomaly)
            except Exception as e:
                logger.error(f"Alert handler {handler.__name__} failed: {e}")

        # Log the alert
        log_level = self._get_log_level(anomaly.severity)
        logger.log(
            log_level,
            f"Anomaly detected: {anomaly.type} for {anomaly.function_id} - "
            f"{anomaly.description}"
        )

    def _get_log_level(self, severity: str) -> int:
        """Get logging level for severity.

        Args:
            severity: The severity level

        Returns:
            Logging level constant
        """
        import logging

        levels = {
            "low": logging.WARNING,
            "medium": logging.ERROR,
            "high": logging.ERROR,
            "critical": logging.CRITICAL,
        }
        return levels.get(severity, logging.WARNING)

    async def check_and_alert(
        self,
        function_id: str,
    ) -> list[AnomalyRecord]:
        """Check for anomalies and generate alerts.

        Args:
            function_id: The function ID to check

        Returns:
            List of detected anomalies
        """
        anomalies = await self._detector.check_all_anomalies(function_id)

        # Generate alerts for new anomalies
        for anomaly in anomalies:
            await self.handle_anomaly(anomaly)

        return anomalies

    async def get_active_alerts(
        self,
        function_id: Optional[str] = None,
    ) -> list[AnomalyRecord]:
        """Get active (unacknowledged) alerts.

        Args:
            function_id: Optional function ID to filter by

        Returns:
            List of active alerts
        """
        all_anomalies = await self._detector.get_anomalies(
            function_id=function_id,
            limit=100,
        )

        # Filter for unacknowledged
        return [a for a in all_anomalies if not a.acknowledged]

    async def acknowledge_alert(
        self,
        anomaly_id: str,
        acknowledged_by: str,
    ) -> bool:
        """Acknowledge an alert.

        Args:
            anomaly_id: The anomaly ID
            acknowledged_by: Who is acknowledging

        Returns:
            True if successful
        """
        return await self._detector.acknowledge_anomaly(anomaly_id, acknowledged_by)


# Default console alert handler
async def console_alert_handler(anomaly: AnomalyRecord) -> None:
    """Default alert handler that prints to console.

    Args:
        anomaly: The detected anomaly
    """
    print(f"\n{'='*60}")
    print(f"🚨 ANOMALY ALERT: {anomaly.type}")
    print(f"{'='*60}")
    print(f"Function: {anomaly.function_id}")
    print(f"Severity: {anomaly.severity.upper()}")
    print(f"Time: {anomaly.detected_at.isoformat()}")
    print(f"Description: {anomaly.description}")
    if anomaly.z_score is not None:
        print(f"Z-Score: {anomaly.z_score:.2f}")
    print(f"{'='*60}\n")


# Global alerting service instance
_alerting_service: Optional[AlertingService] = None


def get_alerting_service() -> AlertingService:
    """Get the global alerting service instance.

    Returns:
        The AlertingService instance
    """
    global _alerting_service
    if _alerting_service is None:
        _alerting_service = AlertingService()
        # Register default handler
        _alerting_service.register_handler(console_alert_handler)
    return _alerting_service
