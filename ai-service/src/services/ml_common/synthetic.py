"""Synthetic data generation for bootstrapping ML models."""

import logging
from typing import List, Tuple

import numpy as np

logger = logging.getLogger(__name__)


def generate_synthetic_request_series(
    hours: int = 168,
    pattern: str = "business_hours",
    base_rate: float = 10.0,
    noise_std: float = 2.0,
    seed: int = 42,
) -> np.ndarray:
    """Generate synthetic request rate time series.

    Args:
        hours: Number of hours to generate
        pattern: One of 'constant', 'business_hours', 'bursty', 'growing'
        base_rate: Base request rate per hour
        noise_std: Standard deviation of noise
        seed: Random seed for reproducibility

    Returns:
        numpy array of hourly request counts
    """
    rng = np.random.RandomState(seed)
    t = np.arange(hours)
    noise = rng.normal(0, noise_std, hours)

    if pattern == "constant":
        series = np.full(hours, base_rate) + noise

    elif pattern == "business_hours":
        hour_of_day = t % 24
        day_of_week = (t // 24) % 7
        hourly_factor = np.where(
            (hour_of_day >= 9) & (hour_of_day <= 17), 2.0, 0.3
        )
        weekday_factor = np.where(day_of_week < 5, 1.0, 0.4)
        series = base_rate * hourly_factor * weekday_factor + noise

    elif pattern == "bursty":
        series = np.full(hours, base_rate * 0.5) + noise
        burst_times = rng.choice(hours, size=hours // 24, replace=False)
        series[burst_times] += base_rate * rng.uniform(3, 8, len(burst_times))

    elif pattern == "growing":
        trend = np.linspace(0.5, 2.0, hours)
        series = base_rate * trend + noise

    else:
        series = np.full(hours, base_rate) + noise

    return np.maximum(series, 0)


def generate_synthetic_cost_series(
    hours: int = 168,
    base_cost: float = 0.005,
    anomaly_count: int = 3,
    seed: int = 42,
) -> Tuple[np.ndarray, np.ndarray]:
    """Generate synthetic cost data with known anomalies.

    Returns:
        Tuple of (cost_series, anomaly_mask) where anomaly_mask is 1 for anomalies
    """
    rng = np.random.RandomState(seed)
    costs = rng.normal(base_cost, base_cost * 0.2, hours)
    costs = np.maximum(costs, 0.0001)

    anomaly_mask = np.zeros(hours, dtype=int)
    if anomaly_count > 0:
        anomaly_indices = rng.choice(
            range(hours // 2, hours), size=anomaly_count, replace=False
        )
        for idx in anomaly_indices:
            costs[idx] = base_cost * rng.uniform(5, 15)
            anomaly_mask[idx] = 1

    return costs, anomaly_mask


def generate_synthetic_latency_series(
    hours: int = 168,
    base_latency: float = 50.0,
    seed: int = 42,
) -> np.ndarray:
    """Generate synthetic latency time series."""
    rng = np.random.RandomState(seed)
    return np.maximum(rng.normal(base_latency, base_latency * 0.15, hours), 1.0)


def generate_synthetic_interactions(
    n_users: int = 50,
    n_functions: int = 100,
    density: float = 0.05,
    seed: int = 42,
) -> np.ndarray:
    """Generate synthetic user-function interaction matrix.

    Args:
        n_users: Number of users
        n_functions: Number of functions
        density: Fraction of non-zero entries
        seed: Random seed

    Returns:
        Sparse interaction matrix (n_users x n_functions) with 0/1 values
    """
    rng = np.random.RandomState(seed)
    matrix = np.zeros((n_users, n_functions))

    for user_idx in range(n_users):
        n_interactions = max(1, int(n_functions * density * rng.uniform(0.5, 1.5)))
        interacted = rng.choice(n_functions, size=min(n_interactions, n_functions), replace=False)
        matrix[user_idx, interacted] = 1.0

    return matrix
