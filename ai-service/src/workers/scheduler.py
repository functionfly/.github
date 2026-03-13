"""Background task scheduler for FlyMind AI Service.

This module provides a scheduler for running periodic background tasks.
"""

import asyncio
import logging
import threading
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, List, Optional
import uuid

logger = logging.getLogger(__name__)


class TaskStatus(str, Enum):
    """Task status."""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class Task:
    """A background task."""
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
    name: str = ""
    func: Optional[Callable] = None
    interval_seconds: int = 60  # 0 means one-time
    enabled: bool = True
    last_run: Optional[datetime] = None
    next_run: Optional[datetime] = None
    run_count: int = 0
    error_count: int = 0

    def should_run(self) -> bool:
        """Check if the task should run."""
        if not self.enabled:
            return False

        if self.interval_seconds == 0:
            # One-time task
            return self.run_count == 0

        if self.next_run is None:
            return True

        return datetime.utcnow() >= self.next_run


@dataclass
class TaskResult:
    """Result of a task execution."""
    task_id: str
    task_name: str
    status: TaskStatus
    started_at: datetime
    completed_at: Optional[datetime] = None
    duration_ms: float = 0.0
    error: Optional[str] = None
    result: Optional[Any] = None

    @property
    def success(self) -> bool:
        """Check if the task succeeded."""
        return self.status == TaskStatus.COMPLETED


class TaskScheduler:
    """Background task scheduler."""

    def __init__(self):
        """Initialize the task scheduler."""
        self._logger = logging.getLogger(__name__)
        self._tasks: Dict[str, Task] = {}
        self._results: List[TaskResult] = []
        self._max_results = 1000
        self._running = False
        self._thread: Optional[threading.Thread] = None
        self._lock = threading.Lock()
        self._stop_event = threading.Event()

    def add_task(
        self,
        name: str,
        func: Callable,
        interval_seconds: int = 60,
    ) -> str:
        """Add a task to the scheduler.

        Args:
            name: Task name
            func: Task function (async or sync)
            interval_seconds: Interval in seconds (0 for one-time)

        Returns:
            Task ID
        """
        with self._lock:
            task = Task(
                name=name,
                func=func,
                interval_seconds=interval_seconds,
                next_run=datetime.utcnow(),
            )

            self._tasks[task.id] = task
            self._logger.info(f"Added task: {name} (interval: {interval_seconds}s)")

            return task.id

    def remove_task(self, task_id: str) -> bool:
        """Remove a task from the scheduler.

        Args:
            task_id: Task ID

        Returns:
            True if removed
        """
        with self._lock:
            if task_id in self._tasks:
                del self._tasks[task_id]
                return True
            return False

    def get_task(self, task_id: str) -> Optional[Task]:
        """Get a task.

        Args:
            task_id: Task ID

        Returns:
            Task or None
        """
        return self._tasks.get(task_id)

    def list_tasks(self) -> List[Task]:
        """List all tasks.

        Returns:
            List of tasks
        """
        return list(self._tasks.values())

    def enable_task(self, task_id: str) -> bool:
        """Enable a task.

        Args:
            task_id: Task ID

        Returns:
            True if enabled
        """
        task = self._tasks.get(task_id)
        if task:
            task.enabled = True
            return True
        return False

    def disable_task(self, task_id: str) -> bool:
        """Disable a task.

        Args:
            task_id: Task ID

        Returns:
            True if disabled
        """
        task = self._tasks.get(task_id)
        if task:
            task.enabled = False
            return True
        return False

    async def run_task(self, task: Task) -> TaskResult:
        """Run a single task.

        Args:
            task: Task to run

        Returns:
            TaskResult
        """
        result = TaskResult(
            task_id=task.id,
            task_name=task.name,
            status=TaskStatus.RUNNING,
            started_at=datetime.utcnow(),
        )

        try:
            self._logger.info(f"Running task: {task.name}")

            # Run the task
            if asyncio.iscoroutinefunction(task.func):
                task_result = await task.func()
            else:
                task_result = task.func()

            result.status = TaskStatus.COMPLETED
            result.result = task_result

        except Exception as e:
            result.status = TaskStatus.FAILED
            result.error = str(e)
            task.error_count += 1
            self._logger.error(f"Task {task.name} failed: {e}")

        finally:
            result.completed_at = datetime.utcnow()
            duration = (result.completed_at - result.started_at).total_seconds() * 1000
            result.duration_ms = duration

            task.last_run = datetime.utcnow()
            task.run_count += 1

            # Schedule next run
            if task.interval_seconds > 0:
                task.next_run = datetime.utcnow()

        return result

    def _run_loop(self) -> None:
        """Main task loop (runs in background thread)."""
        while not self._stop_event.is_set():
            try:
                # Check each task
                for task in list(self._tasks.values()):
                    if task.should_run():
                        # Run the task
                        result = asyncio.run(self.run_task(task))

                        # Store result
                        with self._lock:
                            self._results.append(result)
                            if len(self._results) > self._max_results:
                                self._results = self._results[-self._max_results:]

                # Sleep for a bit
                time.sleep(5)

            except Exception as e:
                self._logger.error(f"Scheduler loop error: {e}")
                time.sleep(5)

    def start(self) -> None:
        """Start the scheduler."""
        if self._running:
            return

        self._running = True
        self._stop_event.clear()

        self._thread = threading.Thread(target=self._run_loop, daemon=True)
        self._thread.start()

        self._logger.info("Task scheduler started")

    def stop(self) -> None:
        """Stop the scheduler."""
        if not self._running:
            return

        self._stop_event.set()

        if self._thread:
            self._thread.join(timeout=10)

        self._running = False
        self._logger.info("Task scheduler stopped")

    def get_results(
        self,
        task_id: Optional[str] = None,
        limit: int = 100,
    ) -> List[TaskResult]:
        """Get task results.

        Args:
            task_id: Optional task ID to filter by
            limit: Maximum results to return

        Returns:
            List of task results
        """
        with self._lock:
            results = self._results

            if task_id:
                results = [r for r in results if r.task_id == task_id]

            return results[-limit:]

    def get_stats(self) -> Dict[str, Any]:
        """Get scheduler statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            total_runs = sum(t.run_count for t in self._tasks.values())
            total_errors = sum(t.error_count for t in self._tasks.values())

            return {
                "running": self._running,
                "total_tasks": len(self._tasks),
                "enabled_tasks": sum(1 for t in self._tasks.values() if t.enabled),
                "total_runs": total_runs,
                "total_errors": total_errors,
                "results_count": len(self._results),
            }


# Global scheduler
_task_scheduler: Optional[TaskScheduler] = None


def get_task_scheduler() -> TaskScheduler:
    """Get the global task scheduler.

    Returns:
        TaskScheduler instance
    """
    global _task_scheduler
    if _task_scheduler is None:
        _task_scheduler = TaskScheduler()

    return _task_scheduler
