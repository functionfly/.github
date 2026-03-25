import os
import time

# Note: This is an example handler meant to be paired with the "kv" capability.
# KV host APIs may vary by runtime; this file exists so the example is publishable
# and can be extended to use KV bindings when available in the Python runtime.

_process_counter = 0


def handler(event):
    global _process_counter
    _process_counter += 1

    key = os.environ.get("COUNTER_KEY", "visit_count")

    return {
        "ok": True,
        "counter_key": key,
        "process_counter": _process_counter,
        "note": "This example increments an in-process counter. Wire to KV bindings for persistence.",
        "ts": int(time.time()),
        "event": event,
    }

