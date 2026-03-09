import time

def handler(event):
    return {"ok": True, "timestamp_ms": int(time.time() * 1000)}
