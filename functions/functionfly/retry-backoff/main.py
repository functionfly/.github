import random
import math


def handler(event):
    attempt = event.get("attempt") if isinstance(event, dict) else None
    base_delay = event.get("base_delay", 1000)
    max_delay = event.get("max_delay", 30000)
    multiplier = event.get("multiplier", 2.0)
    jitter = event.get("jitter", True)
    strategy = event.get("strategy", "exponential")

    if attempt is None:
        return {"ok": False, "error": "attempt is required"}
    try:
        attempt = int(attempt)
    except (TypeError, ValueError):
        return {"ok": False, "error": "attempt must be an integer"}
    if attempt < 0:
        return {"ok": False, "error": "attempt must be >= 0"}

    try:
        base_delay = float(base_delay)
        max_delay = float(max_delay)
        multiplier = float(multiplier)
    except (TypeError, ValueError):
        return {"ok": False, "error": "base_delay, max_delay, and multiplier must be numbers"}

    if strategy == "exponential":
        delay = base_delay * (multiplier ** attempt)
    elif strategy == "linear":
        delay = base_delay * (attempt + 1)
    elif strategy == "constant":
        delay = base_delay
    elif strategy == "fibonacci":
        a, b = 0, 1
        for _ in range(attempt):
            a, b = b, a + b
        delay = base_delay * max(b, 1)
    else:
        return {"ok": False, "error": f"Unknown strategy '{strategy}'. Use: exponential, linear, constant, fibonacci"}

    delay = min(delay, max_delay)

    if jitter:
        delay = delay * (0.5 + random.random() * 0.5)

    delay = round(delay)

    return {
        "ok": True,
        "attempt": attempt,
        "delay_ms": delay,
        "strategy": strategy,
        "jitter_applied": jitter,
    }
