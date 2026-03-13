"""Error analyzer for debugging service.

Analyzes errors and logs to identify root causes.
"""

import logging
import re
from typing import Optional, Dict, Any, List, Tuple
from enum import Enum

from .context_collector import ContextCollector, get_context_collector

logger = logging.getLogger(__name__)


class ErrorCategory(str, Enum):
    """Categories of errors."""
    RUNTIME = "runtime"
    TIMEOUT = "timeout"
    MEMORY = "memory"
    NETWORK = "network"
    AUTHENTICATION = "authentication"
    VALIDATION = "validation"
    RESOURCE = "resource"
    UNKNOWN = "unknown"


# Known error patterns for quick identification
ERROR_PATTERNS = {
    ErrorCategory.TIMEOUT: [
        r"timeout",
        r"timed out",
        r"execution time exceeded",
        r"deadline exceeded",
    ],
    ErrorCategory.MEMORY: [
        r"out of memory",
        r"memory limit",
        r"oom",
        r"memory exceeded",
    ],
    ErrorCategory.NETWORK: [
        r"connection refused",
        r"connection timeout",
        r"network error",
        r"dns lookup failed",
        r"unable to fetch",
        r"http error",
    ],
    ErrorCategory.AUTHENTICATION: [
        r"unauthorized",
        r"authentication failed",
        r"invalid credentials",
        r"token expired",
        r"access denied",
    ],
    ErrorCategory.VALIDATION: [
        r"validation error",
        r"invalid input",
        r"type error",
        r"attribute error",
        r"key error",
        r"value error",
    ],
    ErrorCategory.RESOURCE: [
        r"too many requests",
        r"rate limit",
        r"quota exceeded",
        r"resource not found",
    ],
    ErrorCategory.RUNTIME: [
        r"exception",
        r"traceback",
        r"error:",
        r"failed to",
        r"could not",
    ],
}


class ErrorAnalyzer:
    """Analyzes errors and logs to identify root causes."""

    def __init__(self):
        self._context_collector = get_context_collector()

    async def analyze_error(
        self,
        function_id: str,
        error_message: str,
        stack_trace: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Analyze an error and provide root cause analysis.

        Args:
            function_id: The function ID
            error_message: The error message
            stack_trace: Optional stack trace

        Returns:
            Analysis result with root cause and confidence
        """
        # Collect context
        context = await self._context_collector.collect_context(
            function_id=function_id,
            error_message=error_message,
            stack_trace=stack_trace,
        )

        # Identify error category
        category, category_confidence = self._identify_category(error_message, stack_trace)

        # Extract error details
        details = self._extract_error_details(error_message, stack_trace)

        # Determine root cause
        root_cause, confidence = self._determine_root_cause(
            category,
            error_message,
            stack_trace,
            context,
        )

        # Build analysis result
        analysis = {
            "function_id": function_id,
            "error_message": error_message,
            "error_category": category.value,
            "category_confidence": category_confidence,
            "root_cause": root_cause,
            "confidence": confidence,
            "details": details,
            "context": {
                "function_info": context.get("function_info"),
                "recent_executions_count": len(context.get("recent_executions", [])),
            },
            "suggestions": self._get_initial_suggestions(category, root_cause),
        }

        # Store error for historical tracking
        await self._context_collector.store_error_for_history(
            function_id,
            {
                "error_message": error_message,
                "category": category.value,
                "root_cause": root_cause,
                "analyzed_at": "now",
            },
        )

        return analysis

    def _identify_category(
        self,
        error_message: str,
        stack_trace: Optional[str],
    ) -> Tuple[ErrorCategory, float]:
        """Identify the error category."""
        combined_text = f"{error_message} {stack_trace or ''}".lower()
        scores = {}

        for category, patterns in ERROR_PATTERNS.items():
            score = 0
            for pattern in patterns:
                if re.search(pattern, combined_text, re.IGNORECASE):
                    score += 1
            if score > 0:
                scores[category] = score / len(patterns)

        if not scores:
            return ErrorCategory.UNKNOWN, 0.0

        best_category = max(scores, key=scores.get)
        confidence = scores[best_category]

        return best_category, confidence

    def _extract_error_details(
        self,
        error_message: str,
        stack_trace: Optional[str],
    ) -> Dict[str, Any]:
        """Extract structured error details."""
        details = {
            "error_type": None,
            "error_module": None,
            "error_line": None,
            "possible_values": [],
        }

        type_match = re.search(r"(\w+Error|\w+Exception):", error_message)
        if type_match:
            details["error_type"] = type_match.group(1)

        if stack_trace:
            line_match = re.search(r'File "([^"]+)", line (\d+)', stack_trace)
            if line_match:
                details["error_module"] = line_match.group(1)
                details["error_line"] = int(line_match.group(2))

            func_match = re.search(r'in (\w+)\(', stack_trace)
            if func_match:
                details["error_function"] = func_match.group(1)

        return details

    def _determine_root_cause(
        self,
        category: ErrorCategory,
        error_message: str,
        stack_trace: Optional[str],
        context: Dict[str, Any],
    ) -> Tuple[str, float]:
        """Determine the root cause of the error."""
        root_causes = {
            ErrorCategory.TIMEOUT: [
                "Function execution time exceeded the configured timeout",
                "Cold start delay caused timeout",
                "External API call took too long",
            ],
            ErrorCategory.MEMORY: [
                "Function exceeded memory allocation",
                "Memory leak in function code",
                "Loading large dataset into memory",
            ],
            ErrorCategory.NETWORK: [
                "External API endpoint unavailable",
                "Network connection failed",
                "DNS resolution failed",
            ],
            ErrorCategory.AUTHENTICATION: [
                "Invalid or expired authentication token",
                "Missing required credentials",
                "Insufficient permissions for the operation",
            ],
            ErrorCategory.VALIDATION: [
                "Invalid input data format",
                "Missing required parameters",
                "Type mismatch in function arguments",
            ],
            ErrorCategory.RESOURCE: [
                "Rate limit exceeded for external service",
                "Quota exceeded for cloud resources",
            ],
            ErrorCategory.RUNTIME: [
                "Unhandled exception in function code",
                "Import or dependency error",
                "Logic error in function implementation",
            ],
            ErrorCategory.UNKNOWN: [
                "Unexpected error occurred",
                "Insufficient information to determine root cause",
            ],
        }

        causes = root_causes.get(category, root_causes[ErrorCategory.UNKNOWN])
        confidence = 0.7

        recent_errors = context.get("historical_errors", [])
        if recent_errors:
            for err in recent_errors:
                if err.get("error_message") == error_message:
                    confidence = 0.9
                    break

        return causes[0], confidence

    def _get_initial_suggestions(
        self,
        category: ErrorCategory,
        root_cause: str,
    ) -> List[str]:
        """Get initial suggestions based on category."""
        suggestions_map = {
            ErrorCategory.TIMEOUT: [
                "Increase function timeout if possible",
                "Implement caching for expensive operations",
            ],
            ErrorCategory.MEMORY: [
                "Reduce memory usage in function",
                "Process data in chunks instead of all at once",
            ],
            ErrorCategory.NETWORK: [
                "Add retry logic with exponential backoff",
                "Implement circuit breaker pattern",
            ],
            ErrorCategory.AUTHENTICATION: [
                "Refresh authentication tokens before expiry",
                "Store credentials securely in secrets",
            ],
            ErrorCategory.VALIDATION: [
                "Add input validation at function entry",
                "Use schema validation libraries",
            ],
            ErrorCategory.RESOURCE: [
                "Implement rate limiting in your code",
                "Add caching to reduce API calls",
            ],
            ErrorCategory.RUNTIME: [
                "Add proper error handling and try/catch blocks",
                "Check dependencies for compatibility",
            ],
            ErrorCategory.UNKNOWN: [
                "Review function logs for more details",
            ],
        }

        return suggestions_map.get(category, suggestions_map[ErrorCategory.UNKNOWN])


_error_analyzer: Optional[ErrorAnalyzer] = None


def get_error_analyzer() -> ErrorAnalyzer:
    """Get the global error analyzer instance."""
    global _error_analyzer
    if _error_analyzer is None:
        _error_analyzer = ErrorAnalyzer()
    return _error_analyzer
