"""Atlas trace decorator for automatic function tracing.

This module provides a decorator to automatically trace function execution with Atlas.
"""

import asyncio
import functools
import json
import time
from typing import Any, Callable, Optional

from ..integrations.atlas import AtlasIntegration, CostData, get_atlas_integration

logger = __import__('logging').getLogger(__name__)


def atlas_trace(
    agent_id: str,
    include_args: bool = True,
    include_result: bool = True,
    record_cost: bool = True,
    span_name: Optional[str] = None,
):
    """Decorator to automatically trace function execution with Atlas.

    Args:
        agent_id: Agent identifier
        include_args: Whether to include function arguments
        include_result: Whether to include function results
        record_cost: Whether to record cost
        span_name: Optional span name override

    Returns:
        Decorated function

    Example:
        @atlas_trace("my-agent")
        async def my_function(arg1, arg2):
            return result
    """
    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            atlas = await get_atlas_integration()
            span_id = span_name or func.__name__

            original_span_id = atlas._current_span_id
            if not atlas._current_span_id:
                await atlas.start_span(span_id, {
                    "function": func.__name__,
                    "module": func.__module__,
                })

            start_time = time.time()
            error = None
            result = None

            try:
                if include_args:
                    input_data = {
                        "function": func.__name__,
                        "args": str(args)[:500] if args else "",
                        "kwargs": {k: str(v)[:500] for k, v in list(kwargs.items())[:10]},
                    }
                    await atlas.record_input(json.dumps(input_data))

                result = await func(*args, **kwargs)

                if include_result:
                    result_str = str(result)[:1000] if result else ""
                    await atlas.record_result(result_str)

                return result

            except Exception as e:
                error = e

                await atlas.record_error(str(e), {
                    "function": func.__name__,
                    "args": str(args)[:500] if args else "",
                })

                raise

            finally:
                if record_cost:
                    duration_ms = int((time.time() - start_time) * 1000)
                    cost = CostData(
                        provider="unknown",
                        model="unknown",
                        latency_ms=duration_ms,
                    )

                if not original_span_id:
                    await atlas.end_span(status="completed" if not error else "failed")

        return wrapper
    return decorator


def get_atlas() -> Optional[AtlasIntegration]:
    """Get the global Atlas integration instance synchronously.

    Returns:
        AtlasIntegration instance or None
    """
    try:
        loop = asyncio.get_event_loop()
        if loop.is_running():
            import asyncio
            async def get():
                return await get_atlas_integration()
            from functools import partial
            return partial(asyncio.ensure_future, get())
        else:
            return asyncio.run(get_atlas_integration())
    except Exception as e:
        logger.warning(f"Failed to get Atlas integration: {e}")
        return None