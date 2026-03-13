"""Background Workers package for FlyMind AI Service.

This module provides:
- Background task scheduler
- Periodic tasks (cache warming, anomaly checks)
"""

from .scheduler import (
    TaskScheduler,
    Task,
    TaskResult,
    get_task_scheduler,
)
from .tasks import (
    CacheWarmingTask,
    AnomalyCheckTask,
    MetricsCollectionTask,
    register_default_tasks,
)

__all__ = [
    "TaskScheduler",
    "Task",
    "TaskResult",
    "get_task_scheduler",
    "CacheWarmingTask",
    "AnomalyCheckTask",
    "MetricsCollectionTask",
    "register_default_tasks",
]
