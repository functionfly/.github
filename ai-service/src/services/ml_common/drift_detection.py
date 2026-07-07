"""Model drift detection for FlyMind ML services.

Detects when model input distributions or predictions change significantly,
indicating potential model degradation or data drift.
"""

import json
import logging
import threading
import time
from collections import deque
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional

import numpy as np

from ...config import settings

logger = logging.getLogger(__name__)


@dataclass
class DriftMetrics:
    """Metrics for drift detection."""
    timestamp: datetime
    value: float
    expected_mean: float
    expected_std: float
    z_score: float
    is_drifted: bool = False


@dataclass
class DriftAlert:
    """Drift alert payload."""
    service: str
    tenant_id: str
    metric_name: str
    severity: str  # low, medium, high, critical
    drift_score: float
    drift_percentage: float
    expected_value: float
    actual_value: float
    triggered_at: datetime
    recommendation: str


class DriftDetector:
    """Detects model/data drift using statistical methods.

    Uses Population Stability Index (PSI) and Z-score based detection
    to identify when distributions change significantly.
    """

    PSI_THRESHOLD_LOW = 0.1
    PSI_THRESHOLD_MEDIUM = 0.2
    PSI_THRESHOLD_HIGH = 0.25

    def __init__(self, service: str, tenant_id: str, metric_name: str, window_size: int = 1000):
        """Initialize drift detector.

        Args:
            service: Service name (e.g., 'cost_anomaly', 'recommendations')
            tenant_id: Tenant ID
            metric_name: Name of the metric being tracked
            window_size: Number of samples to keep in rolling window
        """
        self._service = service
        self._tenant_id = tenant_id
        self._metric_name = metric_name
        self._window_size = window_size
        self._lock = threading.Lock()

        self._baseline_values: deque = deque(maxlen=window_size)
        self._current_values: deque = deque(maxlen=window_size)
        self._metrics_history: List[DriftMetrics] = []

        self._baseline_mean: float = 0.0
        self._baseline_std: float = 1.0
        self._baseline_n: int = 0

        self._is_baseline_established = False

    def record_value(self, value: float) -> Optional[DriftMetrics]:
        """Record a value and check for drift.

        Args:
            value: The value to record

        Returns:
            DriftMetrics if drift is detected, None otherwise
        """
        with self._lock:
            self._current_values.append(value)

            if not self._is_baseline_established:
                self._baseline_values.append(value)
                if len(self._baseline_values) >= min(100, self._window_size // 10):
                    self._establish_baseline()
                return None

            if len(self._current_values) < 10:
                return None

            current_mean = np.mean(self._current_values)
            current_std = max(np.std(self._current_values), 0.001)

            z_score = abs(value - self._baseline_mean) / self._baseline_std

            is_drifted = z_score > settings.ml_drift_threshold * 6

            metrics = DriftMetrics(
                timestamp=datetime.utcnow(),
                value=value,
                expected_mean=self._baseline_mean,
                expected_std=self._baseline_std,
                z_score=z_score,
                is_drifted=is_drifted,
            )
            self._metrics_history.append(metrics)

            if len(self._metrics_history) > self._window_size:
                self._metrics_history.pop(0)

            if is_drifted:
                return metrics

            return None

    def _establish_baseline(self) -> None:
        """Establish baseline statistics from collected values."""
        self._baseline_values = deque(self._current_values)
        self._baseline_mean = np.mean(self._baseline_values)
        self._baseline_std = max(np.std(self._baseline_values), 0.001)
        self._baseline_n = len(self._baseline_values)
        self._is_baseline_established = True
        logger.info(
            f"Baseline established for {self._service}/{self._tenant_id}/{self._metric_name}: "
            f"mean={self._baseline_mean:.4f}, std={self._baseline_std:.4f}"
        )

    def compute_psi(self, expected: Optional[np.ndarray] = None, actual: Optional[np.ndarray] = None) -> float:
        """Compute Population Stability Index.

        PSI < 0.1: No significant drift
        0.1 <= PSI < 0.2: Minor drift, monitor
        0.2 <= PSI < 0.25: Moderate drift, investigate
        PSI >= 0.25: Significant drift, action required

        Args:
            expected: Expected distribution (uses baseline if None)
            actual: Actual distribution (uses current if None)

        Returns:
            PSI value
        """
        with self._lock:
            if expected is None:
                expected = np.array(self._baseline_values)
            if actual is None:
                actual = np.array(self._current_values)

            if len(expected) < 10 or len(actual) < 10:
                return 0.0

            min_val = min(expected.min(), actual.min())
            max_val = max(expected.max(), actual.max())

            bins = 10
            bin_edges = np.linspace(min_val, max_val, bins + 1)

            expected_counts, _ = np.histogram(expected, bins=bin_edges)
            actual_counts, _ = np.histogram(actual, bins=bin_edges)

            expected_pct = (expected_counts + 0.5) / (expected_counts.sum() + 0.5 * bins)
            actual_pct = (actual_counts + 0.5) / (actual_counts.sum() + 0.5 * bins)

            psi = np.sum((actual_pct - expected_pct) * np.log(actual_pct / expected_pct))
            return float(psi)

    def get_drift_score(self) -> float:
        """Get current drift score (0.0-1.0).

        Returns:
            Drift score where 0 = no drift, 1 = maximum drift
        """
        with self._lock:
            if not self._is_baseline_established or len(self._current_values) < 10:
                return 0.0

            psi = self.compute_psi()
            return min(1.0, psi / self.PSI_THRESHOLD_HIGH)

    def get_severity(self) -> str:
        """Get current drift severity level.

        Returns:
            One of: 'none', 'low', 'medium', 'high', 'critical'
        """
        score = self.get_drift_score()
        if score < 0.1:
            return "none"
        elif score < 0.3:
            return "low"
        elif score < 0.5:
            return "medium"
        elif score < 0.75:
            return "high"
        else:
            return "critical"

    def get_metrics_summary(self) -> Dict[str, Any]:
        """Get summary of drift metrics.

        Returns:
            Dict with drift detection summary
        """
        with self._lock:
            return {
                "service": self._service,
                "tenant_id": self._tenant_id,
                "metric_name": self._metric_name,
                "is_baseline_established": self._is_baseline_established,
                "baseline_mean": round(self._baseline_mean, 4),
                "baseline_std": round(self._baseline_std, 4),
                "current_sample_count": len(self._current_values),
                "drift_score": round(self.get_drift_score(), 3),
                "severity": self.get_severity(),
                "psi": round(self.compute_psi(), 4),
                "drifted_count": sum(1 for m in self._metrics_history if m.is_drifted),
            }


class DriftMonitor:
    """Monitors drift across all ML services and tenants.

    Maintains drift detectors per service/tenant/metric combination
    and sends alerts when drift is detected.
    """

    def __init__(self):
        self._detectors: Dict[str, DriftDetector] = {}
        self._lock = threading.Lock()
        self._alert_callbacks: List[Callable[[DriftAlert], None]] = []

    def _make_key(self, service: str, tenant_id: str, metric_name: str) -> str:
        """Create a unique key for a detector."""
        return f"{service}:{tenant_id}:{metric_name}"

    def get_or_create_detector(
        self, service: str, tenant_id: str, metric_name: str
    ) -> DriftDetector:
        """Get or create a drift detector.

        Args:
            service: Service name
            tenant_id: Tenant ID
            metric_name: Metric name

        Returns:
            DriftDetector instance
        """
        key = self._make_key(service, tenant_id, metric_name)
        with self._lock:
            if key not in self._detectors:
                self._detectors[key] = DriftDetector(service, tenant_id, metric_name)
            return self._detectors[key]

    def record_metric(
        self,
        service: str,
        tenant_id: str,
        metric_name: str,
        value: float,
    ) -> Optional[DriftAlert]:
        """Record a metric value and check for drift.

        Args:
            service: Service name
            tenant_id: Tenant ID
            metric_name: Metric name
            value: Value to record

        Returns:
            DriftAlert if drift detected and alert triggered, None otherwise
        """
        if not settings.ml_drift_detection_enabled:
            return None

        detector = self.get_or_create_detector(service, tenant_id, metric_name)
        metrics = detector.record_value(value)

        if metrics and metrics.is_drifted:
            alert = self._create_alert(detector, metrics)
            self._trigger_alerts(alert)
            return alert

        return None

    def _create_alert(self, detector: DriftDetector, metrics: DriftMetrics) -> DriftAlert:
        """Create a drift alert."""
        severity = detector.get_severity()
        drift_score = detector.get_drift_score()

        recommendations = {
            "none": "Continue normal monitoring",
            "low": "Monitor but no immediate action needed",
            "medium": "Investigate potential causes of drift",
            "high": "Consider model retraining soon",
            "critical": "Immediate model retraining recommended",
        }

        return DriftAlert(
            service=detector._service,
            tenant_id=detector._tenant_id,
            metric_name=detector._metric_name,
            severity=severity,
            drift_score=drift_score,
            drift_percentage=metrics.z_score * 100 / 6,
            expected_value=metrics.expected_mean,
            actual_value=metrics.value,
            triggered_at=datetime.utcnow(),
            recommendation=recommendations.get(severity, "Unknown"),
        )

    def _trigger_alerts(self, alert: DriftAlert) -> None:
        """Trigger all registered alert callbacks."""
        for callback in self._alert_callbacks:
            try:
                callback(alert)
            except Exception as e:
                logger.error(f"Alert callback failed: {e}")

    def register_alert_callback(self, callback: Callable[[DriftAlert], None]) -> None:
        """Register a callback to be called when drift is detected.

        Args:
            callback: Function to call with DriftAlert
        """
        self._alert_callbacks.append(callback)

    def get_all_drift_status(self) -> List[Dict[str, Any]]:
        """Get drift status for all detectors.

        Returns:
            List of drift status summaries
        """
        with self._lock:
            return [
                detector.get_metrics_summary()
                for detector in self._detectors.values()
            ]

    def get_drift_summary(self) -> Dict[str, Any]:
        """Get overall drift summary.

        Returns:
            Dict with overall drift statistics
        """
        with self._lock:
            statuses = [d.get_metrics_summary() for d in self._detectors.values()]

            by_severity = {"none": 0, "low": 0, "medium": 0, "high": 0, "critical": 0}
            for s in statuses:
                by_severity[s["severity"]] = by_severity.get(s["severity"], 0) + 1

            return {
                "total_detectors": len(statuses),
                "drifted_count": sum(1 for s in statuses if s["severity"] != "none"),
                "by_severity": by_severity,
                "services_monitored": list(set(s["service"] for s in statuses)),
            }


_async_drift_monitor: Optional[DriftMonitor] = None


def get_drift_monitor() -> DriftMonitor:
    """Get the global drift monitor instance."""
    global _async_drift_monitor
    if _async_drift_monitor is None:
        _async_drift_monitor = DriftMonitor()
    return _async_drift_monitor


async def send_drift_alert_webhook(alert: DriftAlert) -> bool:
    """Send a drift alert to a webhook URL.

    Args:
        alert: The drift alert to send

    Returns:
        True if alert was sent successfully
    """
    import httpx

    webhook_url = settings.ml_drift_alert_webhook_url
    if not webhook_url:
        return False

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.post(
                webhook_url,
                json={
                    "event": "ml_drift_alert",
                    "service": alert.service,
                    "tenant_id": alert.tenant_id,
                    "metric_name": alert.metric_name,
                    "severity": alert.severity,
                    "drift_score": alert.drift_score,
                    "drift_percentage": alert.drift_percentage,
                    "expected_value": alert.expected_value,
                    "actual_value": alert.actual_value,
                    "triggered_at": alert.triggered_at.isoformat(),
                    "recommendation": alert.recommendation,
                },
            )
            return response.status_code < 400
    except Exception as e:
        logger.error(f"Failed to send drift alert webhook: {e}")
        return False
