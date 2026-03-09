import time


def handler(event):
    if isinstance(event, dict):
        key = event.get("key", "")
        limit = event.get("limit", 100)
        window_seconds = event.get("window_seconds", 60)
        current_count = event.get("current_count", 0)
        window_start = event.get("window_start")
        now = event.get("now")
    else:
        key, limit, window_seconds, current_count, window_start, now = "", 100, 60, 0, None, None

    if not key:
        return {"ok": False, "error": "Input 'key' is required"}
    try:
        limit = max(1, int(limit))
        window_seconds = max(1, int(window_seconds))
        current_count = max(0, int(current_count))
    except (TypeError, ValueError):
        return {"ok": False, "error": "limit, window_seconds, current_count must be integers"}

    now = now if now is not None else time.time()
    if window_start is None:
        window_start = now

    elapsed = now - window_start
    if elapsed >= window_seconds:
        remaining = limit - 1
        allowed = True
        reset_in = window_seconds
    else:
        remaining = max(0, limit - current_count - 1)
        allowed = current_count < limit
        reset_in = max(0, window_seconds - elapsed)

    return {"ok": True, "allowed": allowed, "remaining": remaining, "reset_in_seconds": round(reset_in, 2)}
