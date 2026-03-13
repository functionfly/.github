"""Structured logging for FlyMind AI Service.

This module provides structured JSON logging with context.
"""

import json
import logging
import sys
import threading
from datetime import datetime
from enum import Enum
from typing import Any, Dict, Optional
from dataclasses import dataclass, field


class LogLevel(str, Enum):
    """Log levels."""
    DEBUG = "debug"
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


@dataclass
class LogContext:
    """Context for structured logging."""
    tenant_id: Optional[str] = None
    user_id: Optional[str] = None
    request_id: Optional[str] = None
    function_id: Optional[str] = None
    session_id: Optional[str] = None
    extra: Dict[str, Any] = field(default_factory=dict)


class StructuredLogger:
    """Structured JSON logger."""

    def __init__(
        self,
        name: str,
        context: Optional[LogContext] = None,
    ):
        """Initialize the structured logger.

        Args:
            name: Logger name
            context: Default log context
        """
        self._logger = logging.getLogger(name)
        self._context = context or LogContext()
        self._lock = threading.Lock()

    def _format_message(
        self,
        level: str,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> str:
        """Format a structured log message.

        Args:
            level: Log level
            message: Log message
            context: Log context
            **kwargs: Additional fields

        Returns:
            JSON formatted string
        """
        ctx = context or self._context

        log_data = {
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "level": level,
            "logger": self._logger.name,
            "message": message,
        }

        # Add context
        if ctx.tenant_id:
            log_data["tenant_id"] = ctx.tenant_id
        if ctx.user_id:
            log_data["user_id"] = ctx.user_id
        if ctx.request_id:
            log_data["request_id"] = ctx.request_id
        if ctx.function_id:
            log_data["function_id"] = ctx.function_id
        if ctx.session_id:
            log_data["session_id"] = ctx.session_id

        # Add extra fields
        log_data.update(ctx.extra)
        log_data.update(kwargs)

        return json.dumps(log_data)

    def debug(
        self,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> None:
        """Log a debug message."""
        msg = self._format_message("debug", message, context, **kwargs)
        self._logger.debug(msg)

    def info(
        self,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> None:
        """Log an info message."""
        msg = self._format_message("info", message, context, **kwargs)
        self._logger.info(msg)

    def warning(
        self,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> None:
        """Log a warning message."""
        msg = self._format_message("warning", message, context, **kwargs)
        self._logger.warning(msg)

    def error(
        self,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> None:
        """Log an error message."""
        msg = self._format_message("error", message, context, **kwargs)
        self._logger.error(msg)

    def critical(
        self,
        message: str,
        context: Optional[LogContext] = None,
        **kwargs,
    ) -> None:
        """Log a critical message."""
        msg = self._format_message("critical", message, context, **kwargs)
        self._logger.critical(msg)

    def exception(
        self,
        message: str,
        context: Optional[LogContext] = None,
        exc_info: Optional[Exception] = None,
        **kwargs,
    ) -> None:
        """Log an exception."""
        error_data = {
            "exception_type": type(exc_info).__name__ if exc_info else None,
            "exception_message": str(exc_info) if exc_info else None,
        }

        if exc_info and hasattr(exc_info, "__traceback__"):
            import traceback
            error_data["traceback"] = "".join(
                traceback.format_tb(exc_info.__traceback__)
            )

        msg = self._format_message(
            "error",
            message,
            context,
            **error_data,
            **kwargs,
        )
        self._logger.exception(msg)

    def with_context(
        self,
        tenant_id: Optional[str] = None,
        user_id: Optional[str] = None,
        request_id: Optional[str] = None,
        function_id: Optional[str] = None,
        session_id: Optional[str] = None,
        **extra,
    ) -> "StructuredLogger":
        """Create a new logger with additional context.

        Args:
            tenant_id: Tenant ID
            user_id: User ID
            request_id: Request ID
            function_id: Function ID
            session_id: Session ID
            **extra: Extra fields

        Returns:
            New StructuredLogger with context
        """
        new_context = LogContext(
            tenant_id=tenant_id or self._context.tenant_id,
            user_id=user_id or self._context.user_id,
            request_id=request_id or self._context.request_id,
            function_id=function_id or self._context.function_id,
            session_id=session_id or self._context.session_id,
            extra={**self._context.extra, **extra},
        )

        return StructuredLogger(self._logger.name, new_context)


# Global logger cache
_loggers: Dict[str, StructuredLogger] = {}
_logger_lock = threading.Lock()


def get_logger(name: str) -> StructuredLogger:
    """Get a structured logger.

    Args:
        name: Logger name

    Returns:
        StructuredLogger instance
    """
    with _logger_lock:
        if name not in _loggers:
            _loggers[name] = StructuredLogger(name)
        return _loggers[name]


def setup_logging(
    level: str = "INFO",
    format_json: bool = True,
    include_caller: bool = False,
) -> None:
    """Setup logging configuration.

    Args:
        level: Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        format_json: Whether to use JSON format
        include_caller: Whether to include caller info
    """
    log_level = getattr(logging, level.upper(), logging.INFO)

    # Create handler
    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(log_level)

    # Set formatter
    if format_json:
        # Use a custom JSON formatter
        class JSONFormatter(logging.Formatter):
            def format(self, record: logging.LogRecord) -> str:
                log_data = {
                    "timestamp": datetime.utcfromtimestamp(record.created).isoformat() + "Z",
                    "level": record.levelname,
                    "logger": record.name,
                    "message": record.getMessage(),
                }

                if include_caller:
                    log_data["caller"] = {
                        "file": record.filename,
                        "line": record.lineno,
                        "function": record.funcName,
                    }

                if record.exc_info:
                    log_data["exception"] = self.formatException(record.exc_info)

                return json.dumps(log_data)

        formatter = JSONFormatter()
    else:
        formatter = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )

    handler.setFormatter(formatter)

    # Configure root logger
    root_logger = logging.getLogger()
    root_logger.setLevel(log_level)
    root_logger.handlers.clear()
    root_logger.addHandler(handler)

    # Set third-party loggers to WARNING
    logging.getLogger("uvicorn").setLevel(logging.WARNING)
    logging.getLogger("fastapi").setLevel(logging.WARNING)
    logging.getLogger("httpx").setLevel(logging.WARNING)
